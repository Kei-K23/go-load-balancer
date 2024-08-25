package main

import (
	"log"
	"net/http"

	loadbalancer "github.com/Kei-K23/go-load-balancer/load-balancer"
)

const (
	SERVER_PORT = ":8080"
)

// This is main entry point
func main() {
	// Init connection server pool
	serverPool := loadbalancer.NewServerPool()

	// Add backend servers
	serverPool.AddServer(&loadbalancer.Server{Address: "http://localhost:8081", Alive: true})
	serverPool.AddServer(&loadbalancer.Server{Address: "http://localhost:8082", Alive: true})
	serverPool.AddServer(&loadbalancer.Server{Address: "http://localhost:8083", Alive: true})
	serverPool.AddServer(&loadbalancer.Server{Address: "http://localhost:8085", Alive: true})

	// Check servers health concurrently
	go serverPool.HealthChecker()

	http.HandleFunc("/", serverPool.ServeHTTP)

	log.Println("Load Balancer started on :8080")

	if err := http.ListenAndServe(SERVER_PORT, nil); err != nil {
		log.Fatal(err)
	}
}
