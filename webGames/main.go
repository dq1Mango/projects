package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
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

	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})

	r.GET("/register", authorizePrime(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})

	r.POST("/register", registerLogic(db))

	r.GET("/login", authorizePrime(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", loginLogic(db))

	r.GET("/dashboard", authorize(), func(c *gin.Context) {

		userId, _ := c.Get("id")
		id := userId.(int)
		username, err := getUserName(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not find username",
			})
		}

		c.HTML(http.StatusOK, "dashboard.tmpl.html", gin.H{"name": username})
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

	//r.GET("/static", func(c *gin.Context) {
	//	g.HTML
	//})

	r.StaticFile("/static", "./svelte/dist/index.html")

	r.Static("/assets", "./svelte/dist/assets")

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
