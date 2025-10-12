package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
			return
		}

		claims, err := decodeToken(token)
		if err != nil {
			unauthorized(c)
			return
		}

		// implement token refresh here

		c.Set("id", claims.Id)

		c.Next()
	}
}

// these functions seems pretty similar but i dont think they can be nicely consolidated
func authorizePrime() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth")
		if err != nil {
			c.Next()
			return
		}

		fmt.Println("decoding token anyway ...")
		_, err = decodeToken(token)
		if err != nil {
			c.Next()
			return
		}

		c.Redirect(http.StatusFound, "/dashboard")
		c.Abort()
		return
	}
}

func registerLogic(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		c.Redirect(http.StatusFound, "/dashboard")

	}
}

func loginLogic(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginUser struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&loginUser); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			fmt.Println(err)
			return
		}

		id, err := getId(db, loginUser.Username)
		if err == sql.ErrNoRows {
			c.JSON(403, gin.H{"error": "User does not exist"})
			return
		}

		hash, err := getHash(db, id)
		if err = bcrypt.CompareHashAndPassword(hash, []byte(loginUser.Password)); err != nil {
			c.JSON(401, gin.H{"error": "Username and password do not match"})
			return
		}

		setJWTCookie(c, id)

		c.JSON(http.StatusOK, gin.H{"redirect": "/dashboard"})

		//c.Redirect(302, "/dashboard")

	}
}
