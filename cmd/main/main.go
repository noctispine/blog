package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
}

func main() {
	fmt.Println("Hello")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello %s", c.Request.Header)
	})
	r.GET("/healthcheck", healthCheckHandler)

	r.Run(":8080")
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
