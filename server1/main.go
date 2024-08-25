package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	SERVER_ADDR = "0.0.0.0:8081"
)

func main() {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("This is health response from %s", SERVER_ADDR),
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"address": fmt.Sprintf("Backend server address: %s", SERVER_ADDR),
			"message": "pong",
		})
	})
	r.Run(SERVER_ADDR)
}
