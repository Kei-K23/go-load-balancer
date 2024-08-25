package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	SERVER_ADDR = "0.0.0.0:8082"
)

func main() {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("This is health response from %s", SERVER_ADDR),
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"address": SERVER_ADDR,
		})
	})
	r.Run(SERVER_ADDR)
}
