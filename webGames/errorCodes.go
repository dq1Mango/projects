package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// error "constants"
var ErrNotInSlice = errors.New("not in slice")
var ErrAlreadyInSlice = errors.New("already in slice")

func unauthorized(c *gin.Context) {
	//c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	c.Redirect(http.StatusSeeOther, "/login")
	c.Abort()
	return
}
