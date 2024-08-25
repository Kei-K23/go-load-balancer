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

	// Start health checker and weight adjustment in the background
	go serverPool.HealthChecker()
	go serverPool.StartWeightAdjustment()

	http.HandleFunc("/", serverPool.ServeHTTP)

	log.Println("Load Balancer started on :8080")

	if err := http.ListenAndServe(SERVER_PORT, nil); err != nil {
		log.Fatal(err)
	}
}
