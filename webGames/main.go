package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func setJWTCookie(c *gin.Context) {
	token, err := genToken()
	if err != nil {
		fmt.Println("Error in generating auth token: ", err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
	}
}

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
		var newUser struct {
			Username        string `json:"username"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirmPassword"`
		}

		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// Validate passwords match
		if newUser.Password != newUser.ConfirmPassword {
			c.JSON(400, gin.H{"error": "Passwords do not match"})
			return
		}

		// check avaliability
		if nameExists(db, newUser.Username) {
			c.JSON(400, gin.H{"error": "Username is taken"})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Server error"})
			return
		}

		// database logic
		makeNewUser(db, newUser.Username, hash)

	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
		//username := c.PostForm("username")
		//password := c.PostForm("password")

	})

	r.GET("/me", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		username := session.Get("username")

		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
