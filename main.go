package main

import "sync"

type Server struct {
	Address     string
	Alive       bool
	Mu          sync.RWMutex
	Connections int
}

func (s *Server) IsAlive() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.Alive
}

func (s *Server) SetAlive(alive bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Alive = alive
}

func main() {

}
