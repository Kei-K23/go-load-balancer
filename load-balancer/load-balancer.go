package loadbalancer

import (
	"context"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	Address           string
	Alive             bool
	Mu                sync.RWMutex
	Connections       int32         // Use atomic operations for connection counter
	ResponseTime      time.Duration // Track server response time
	Weight            float64       // Dynamic weight based on performance
	FailureCount      int           // Track consecutive failures for circuit breaker
	LastFailed        time.Time     // Last failed attempt time
	CircuitBreaker    bool          // Circuit breaker state
	BreakerThreshold  int           // Number of failures before tripping the circuit breaker
	BreakerResetAfter time.Duration // Time after which the breaker resets
}

type ServerPool struct {
	Servers []*Server
	Mu      sync.RWMutex
	Logger  *zap.Logger
}

// NewServer initializes a new server with default values
func NewServer(address string) *Server {
	return &Server{
		Address:           address,
		Alive:             true,
		ResponseTime:      10 * time.Millisecond,
		BreakerThreshold:  5,
		BreakerResetAfter: 1 * time.Minute,
	}
}

// Method to check if the server is alive
func (s *Server) IsAlive() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.Alive && !s.CircuitBreaker
}

// Set server status
func (s *Server) SetAlive(alive bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Alive = alive
}

// Increment the server's connection count
func (s *Server) IncrementConnections() {
	atomic.AddInt32(&s.Connections, 1)
}

// Decrement the server's connection count
func (s *Server) DecrementConnections() {
	atomic.AddInt32(&s.Connections, -1)
}

// Circuit breaker logic to avoid overloading failing servers
func (s *Server) TripCircuitBreaker() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.CircuitBreaker = true
	s.LastFailed = time.Now()
}

// Reset the circuit breaker after a timeout
func (s *Server) ResetCircuitBreaker() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if time.Since(s.LastFailed) > s.BreakerResetAfter {
		s.CircuitBreaker = false
		s.FailureCount = 0
	}
}

func NewServerPool(logger *zap.Logger) *ServerPool {
	return &ServerPool{
		Servers: make([]*Server, 0),
		Logger:  logger,
	}
}

// Add a server to the server pool
func (sp *ServerPool) AddServer(newServer *Server) {
	sp.Mu.Lock()
	defer sp.Mu.Unlock()
	sp.Servers = append(sp.Servers, newServer)
}

// Get the count of available servers
func (sp *ServerPool) GetServerCount() int {
	sp.Mu.RLock()
	defer sp.Mu.RUnlock()
	return len(sp.Servers)
}

// Start background processes for weight adjustment and health checks
func (sp *ServerPool) StartBackgroundTasks(ctx context.Context) {
	go sp.HealthChecker(ctx)
	go sp.StartWeightAdjustment(ctx)
}

// Adjust server weights dynamically based on their performance
func (sp *ServerPool) AdjustWeights() {
	sp.Mu.Lock()
	defer sp.Mu.Unlock()

	minResponseTime := time.Millisecond * 1 // Minimum response time threshold to avoid division by zero

	for _, server := range sp.Servers {
		if server.IsAlive() {

			responseTime := server.ResponseTime
			if responseTime < minResponseTime {
				responseTime = minResponseTime // Prevent division by zero or very small times
			}

			// Calculate weight based on response time and number of connections
			serverWeight := math.Exp(-float64(responseTime.Milliseconds())/1000) / (1.0 + float64(server.Connections))

			server.Weight = serverWeight
		}
	}
}

// Weighted round-robin selection based on server weights
func (sp *ServerPool) WeightedRoundRobin() *Server {
	sp.Mu.RLock()
	defer sp.Mu.RUnlock()

	if len(sp.Servers) == 0 {
		return nil
	}

	totalWeight := 0.0

	for _, server := range sp.Servers {
		if server.IsAlive() {
			totalWeight += server.Weight
		}
	}

	if totalWeight == 0 {
		return nil
	}

	randWeight := rand.Float64() * totalWeight

	for _, server := range sp.Servers {
		if server.IsAlive() {
			randWeight -= server.Weight
			if randWeight <= 0 {
				return server
			}
		}
	}
	return nil
}

// Health checker to monitor server status
func (sp *ServerPool) HealthChecker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, server := range sp.Servers {
				go sp.CheckServerHealth(server)
			}
		}
	}
}

// Check the health of a single server
func (sp *ServerPool) CheckServerHealth(server *Server) {
	resp, err := http.Get(server.Address + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		server.SetAlive(false)
		server.FailureCount++
		if server.FailureCount >= server.BreakerThreshold {
			server.TripCircuitBreaker()
			sp.Logger.Warn("Server tripped circuit breaker", zap.String("server", server.Address))
		}
	} else {
		server.SetAlive(true)
		server.ResetCircuitBreaker()
	}
}

// Start weight adjustment periodically
func (sp *ServerPool) StartWeightAdjustment(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sp.AdjustWeights()
		}
	}
}

// ServeHTTP handles incoming HTTP requests and forwards them to a selected backend server
func (sp *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := sp.WeightedRoundRobin()
	if server == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	server.IncrementConnections()
	defer server.DecrementConnections()

	startTime := time.Now()

	sp.Logger.Info("Forwarding request", zap.String("server", server.Address))

	proxyReq, err := http.NewRequest(r.Method, server.Address+r.RequestURI, r.Body)
	if err != nil {
		http.Error(w, "Server unavailable", http.StatusServiceUnavailable)
		sp.Logger.Error("Failed to create proxy request", zap.Error(err))
		return
	}

	proxyReq.Header = r.Header
	client := &http.Client{}

	resp, err := client.Do(proxyReq)
	if err != nil {
		server.FailureCount++
		if server.FailureCount >= server.BreakerThreshold {
			server.TripCircuitBreaker()
		}
		http.Error(w, "Server unavailable", http.StatusServiceUnavailable)
		sp.Logger.Error("Failed to forward request", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	server.Mu.Lock()
	server.ResponseTime = time.Since(startTime)
	server.Mu.Unlock()

	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
