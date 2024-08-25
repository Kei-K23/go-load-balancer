// Load balancer package include all necessary methods and structs

package loadbalancer

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	Address      string       // Server address e.g http://localhost:8081
	Alive        bool         // Use for health check
	Mu           sync.RWMutex // Read-write mutex to work with concurrency
	Connections  int
	ResponseTime time.Duration // Track server response time
	Weight       float64       // Dynamic weight based on performances
}

type ServerPool struct {
	Servers []*Server
	Current uint
	Mu      sync.RWMutex
}

// Method to check server is active
func (s *Server) IsAlive() bool {
	s.Mu.RLock() // Read lock
	defer s.Mu.RUnlock()
	return s.Alive
}

func (s *Server) SetAlive(alive bool) {
	s.Mu.Lock() // Full write lock
	defer s.Mu.Unlock()
	s.Alive = alive
}

func (s *Server) IncrementConnections() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Connections++
}

func (s *Server) DecrementConnections() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if s.Connections > 0 {
		s.Connections--
	}
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		Servers: make([]*Server, 0),
	}
}

func (sp *ServerPool) AddServer(newServer *Server) {
	sp.Mu.Lock()
	defer sp.Mu.Unlock()
	sp.Servers = append(sp.Servers, newServer)
}

func (sp *ServerPool) GetServerCount() int {
	sp.Mu.RLock()
	defer sp.Mu.RUnlock()
	return len(sp.Servers)
}

// Run the weight adjustment periodically to adapt to changing server conditions.
func (sp *ServerPool) StartWeightAdjustment() {
	for {
		sp.AdjustWeights()
		time.Sleep(10 * time.Second)
	}
}

// Adjust the weights of servers dynamically based on their response time
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
			serverWeight := 1.0 / (float64(responseTime.Milliseconds()) * float64(server.Connections+1))
			server.Weight = serverWeight
			fmt.Printf("Adjusted Weight for %s: %f\n", server.Address, server.Weight)
		}
	}
}

// Dynamic weights to distribute requests more intelligently
func (sp *ServerPool) WeightedRoundRobin() *Server {
	sp.Mu.Lock()
	defer sp.Mu.Unlock()

	if len(sp.Servers) == 0 {
		return nil
	}

	totalWeight := 0.0

	for _, server := range sp.Servers {
		if server.IsAlive() {
			totalWeight += server.Weight
			fmt.Printf("Server %s Weight: %f\n", server.Address, server.Weight)
		}
	}

	if totalWeight == 0 {
		return nil
	}

	randWeight := rand.Float64() * totalWeight
	fmt.Printf("Total Weight: %f, Random Weight: %f\n", totalWeight, randWeight)

	for _, server := range sp.Servers {
		if server.IsAlive() {
			randWeight -= server.Weight
			if randWeight <= 0 {
				fmt.Printf("Selected Server: %s\n", server.Address)
				return server
			}
		}
	}
	return nil
}

// Health checker ensures that only healthy backend servers are used. It periodically checks each server by sending an HTTP request
func (sp *ServerPool) HealthChecker() {
	for {
		if len(sp.Servers) == 0 {
			time.Sleep(10 * time.Second)
			continue
		}

		for _, server := range sp.Servers {
			resp, err := http.Get(server.Address + "/health")
			if err != nil || resp.StatusCode != http.StatusOK {
				server.SetAlive(false)
			} else {
				server.SetAlive(true)
			}
		}
		// Run health check every 10 seconds
		time.Sleep(10 * time.Second)
	}
}

// Handles incoming HTTP requests, selects a backend server using the RoundRobin algorithm, and forwards the request to that server. It acts as a reverse proxy, managing the communication between the client and the backend server.
func (sp *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := sp.WeightedRoundRobin()
	if server == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Increment the connection
	server.IncrementConnections()
	// Decrement the connection when this function execution is out of scope
	defer server.DecrementConnections()

	fmt.Printf("Server Address: %s | Alive: %t | Connections: %d | Server Weight: %f\n", server.Address, server.IsAlive(), server.Connections, server.Weight)

	proxyReq, err := http.NewRequest(r.Method, server.Address+r.RequestURI, r.Body)
	if err != nil {
		http.Error(w, "Server unavailable", http.StatusServiceUnavailable)
		return
	}

	proxyReq.Header = r.Header
	client := &http.Client{}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Server unavailable", http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()
	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
