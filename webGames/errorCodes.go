package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	c.Abort()
	return
}
