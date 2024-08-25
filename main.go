package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	loadbalancer "github.com/Kei-K23/go-load-balancer/load-balancer"
	"go.uber.org/zap"
)

const (
	SERVER_PORT = ":8080"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

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

	// Graceful shutdown handling
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		logger.Info("Shutting down load balancer...")

		cancel()

		// Create a deadline to wait for current requests to complete
		ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutDown()

		if err := server.Shutdown(ctxShutDown); err != nil {
			logger.Fatal("Server Shutdown Failed", zap.Error(err))
		}

		logger.Info("Load Balancer shutdown gracefully")
	}()

	logger.Info("Load Balancer started on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
