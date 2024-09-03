package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func main() {
	r := gin.Default()

	// Open the database connection
	db, err := OpenDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure the test user exists
	err = EnsureTestUser(db)
	if err != nil {
		log.Fatalf("Failed to ensure test user: %v", err)
	}

	// Serve static files and templates
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Apply the session middleware globally
	r.Use(SessionMiddleware())

	// Public routes
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
		LoginHandler(c, OpenDB)
	})

	r.GET("/logout", LogoutHandler)

	// Protected routes (require authentication)
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			session := c.MustGet("session").(*sessions.Session)
			userID := session.Values["user_id"]
			c.HTML(http.StatusOK, "dashboard.html", gin.H{"UserID": userID})
		})
	}

	r.Run(":8080")
}
