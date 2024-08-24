// Load balancer package include all necessary methods and structs

package loadbalancer

import "sync"

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
