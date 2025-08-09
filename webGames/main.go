package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func setJWTCookie(c *gin.Context, id int) {
	token, err := genToken(id)
	if err != nil {
		fmt.Println("Error in generating auth token: ", err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
	}

	c.SetCookie("auth", token, 3600*24*7, "/", "", false, true) // flip second to last arg when i finally figure out https
}

func authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth")
		if err != nil {
			unauthorized(c)
		}

		claims, err := decodeToken(token)
		if err != nil {
			unauthorized(c)
		}

		// implement token refresh here

		c.Set("id", claims.Id)

		c.Next()
	}
}

func authorizePrime() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth")
		if err != nil {
			c.Next()
		}

		_, err = decodeToken(token)
		if err != nil {
			c.Next()
		}

		c.Redirect(http.StatusFound, "/me")
		c.Abort()
		return
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

	r.GET("/register", authorizePrime(), func(c *gin.Context) {
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
		userId, _ := makeNewUser(db, newUser.Username, hash)

		setJWTCookie(c, userId)

		c.Redirect(300, "/me")

	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
		//username := c.PostForm("username")
		//password := c.PostForm("password")

	})

	r.GET("/me", authorize(), func(c *gin.Context) {

		userId, _ := c.Get("id") // it will exists trust
		id := userId.(int)

		username, _ := getUserName(db, id)

		c.JSON(http.StatusOK, gin.H{
			"user_id":  id,
			"username": username,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
