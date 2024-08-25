// Load balancer package include all necessary methods and structs

package loadbalancer

import (
	"hash/fnv"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	Address     string       // Server address e.g http://localhost:8081
	Alive       bool         // Use for health check
	Mu          sync.RWMutex // Read-write mutex to work with concurrency
	Connections int          // To track how many user is connected to the server
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

type ServerPool struct {
	Servers []*Server
	Current uint
	Mu      sync.RWMutex
}

func NewServerPool() *ServerPool {
	return &ServerPool{}
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

// Round-robin algorithm cycles through the list of servers, distributing requests evenly
func (sp *ServerPool) RoundRobin() *Server {
	sp.Mu.Lock()
	defer sp.Mu.Unlock()

	// Get the server according to round-robin algorithm
	server := sp.Servers[sp.Current%uint(len(sp.Servers))]
	// Increment current state of server pool
	sp.Current++

	return server
}

// Least connections algorithm forwards the request to the server with the fewest active connections
func (sp *ServerPool) LeastConnections() *Server {
	sp.Mu.RLock()
	defer sp.Mu.Lock()

	var selectedServer *Server
	minConnections := int(^uint(0) >> 1)

	for _, server := range sp.Servers {
		if server.IsAlive() && server.Connections < minConnections {
			minConnections = server.Connections
			selectedServer = server
		}
	}

	return selectedServer
}

// IP hash algorithm maps client IP addresses to specific backend servers, providing session persistence
func (sp *ServerPool) IPHash(clientIP string) *Server {
	sp.Mu.RLock()
	defer sp.Mu.RUnlock()

	hash := fnv.New32a()
	hash.Write([]byte(clientIP))
	index := hash.Sum32() % uint32(len(sp.Servers))

	return sp.Servers[index]
}

// Health checker ensures that only healthy backend servers are used. It periodically checks each server by sending an HTTP request
func (sp *ServerPool) HealthChecker() {
	for {
		for _, server := range sp.Servers {
			resp, err := http.Get(server.Address + "/health")
			if err != nil && resp.StatusCode != http.StatusOK {
				server.SetAlive(false)
			} else {
				server.SetAlive(true)
			}
		}
		// Run health check for every 10 second
		time.Sleep(10 * time.Second)
	}
}
