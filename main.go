package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	loadbalancer "github.com/Kei-K23/go-load-balancer/load-balancer"
	"github.com/Kei-K23/go-load-balancer/logger"
)

const (
	SERVER_PORT = ":8080"
)

func main() {
	logger := logger.NewLogger()

	serverPool := loadbalancer.NewServerPool(logger)

	serverPool.AddServer(loadbalancer.NewServer("http://localhost:8081"))
	serverPool.AddServer(loadbalancer.NewServer("http://localhost:8082"))
	serverPool.AddServer(loadbalancer.NewServer("http://localhost:8083"))

	ctx, cancel := context.WithCancel(context.Background())

	serverPool.StartBackgroundTasks(ctx)

	server := &http.Server{
		Addr:    SERVER_PORT,
		Handler: http.HandlerFunc(serverPool.ServeHTTP),
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		logger.Info("Shutting down load balancer...")

		cancel()

		ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutDown()

		if err := server.Shutdown(ctxShutDown); err != nil {
			logger.Error("Server Shutdown Failed: %v", err)
		}

		logger.Info("Load Balancer shutdown gracefully")
	}()

	logger.Info("Load Balancer started on %s", SERVER_PORT)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server failed: %v", err)
	}
}
