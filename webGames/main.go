package main

import (
	//	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	db := initDB()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})

	r.POST("/register", func(c *gin.Context) {
		var user struct {
			Username        string `json:"username"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirmPassword"`
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// Validate passwords match
		if user.Password != user.ConfirmPassword {
			c.JSON(400, gin.H{"error": "Passwords do not match"})
			return
		}

		users, err := db.Query("select * from auth where username = \"" + user.Username + "\"")
		if users != nil {
			c.JSON(400, gin.H{"error": "Username is taken"})
			return
		}

		// Hash password
		_, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Server error"})
			return
		}

		// database logic

	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
		//username := c.PostForm("username")
		//password := c.PostForm("password")

	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
