package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"
)

func LoginUser(w http.ResponseWriter, r *http.Request, db *sql.DB, username, password string) (int, error) {
	// Fetch the user by username
	user, err := GetUserByUsername(db, username)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found:", username)
			return 0, errors.New("invalid username or password")
		}
		return 0, err
	}

	fmt.Printf("Retrieved user: %v\n", user) // Debug output

	// Check if the password is correct
	if !CheckPasswordHash(password, user.PasswordHash) {
		fmt.Println("Password mismatch")
		return 0, errors.New("invalid username or password")
	}

	// Set the session with the user ID
	err = SetSession(w, r, "user_id", user.ID)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func LogoutUser(w http.ResponseWriter, r *http.Request) error {
	return ClearSession(w, r)
}

func LoginHandler(c *gin.Context, dbFunc func() (*sql.DB, error)) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	db, err := dbFunc()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer db.Close()

	_, err = LoginUser(c.Writer, c.Request, db, username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Set a custom header to indicate a redirect
	c.Header("HX-Redirect", "/dashboard")
	c.Status(http.StatusOK)
}

func LogoutHandler(c *gin.Context) {
	// Retrieve the session
	session := c.MustGet("session").(*sessions.Session)

	// Clear the session
	session.Options.MaxAge = -1 // This deletes the session cookie
	err := session.Save(c.Request, c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear session"})
		return
	}

	// Render the logout.html template
	c.HTML(http.StatusOK, "logout.html", nil)
}

func DashboardHandler(c *gin.Context) {
	session := c.MustGet("session").(*sessions.Session)

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID from session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Welcome to the dashboard, User %d!", userID),
		"user_id": userID,
	})
}
