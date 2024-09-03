package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

func TestSetSession(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	err := SetSession(w, r, "username", "testuser")
	if err != nil {
		t.Fatalf("SetSession failed: %v", err)
	}

	session, _ := sessionStore.Get(r, "session-name")
	if session.Values["username"] != "testuser" {
		t.Errorf("Expected session value 'testuser', got '%v'", session.Values["username"])
	}
}

func TestGetSession(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	session, _ := sessionStore.Get(r, "session-name")
	session.Values["username"] = "testuser"
	session.Save(r, w)

	cookie := w.Result().Cookies()[0]
	r = httptest.NewRequest("GET", "/", nil)
	r.AddCookie(cookie)

	value, err := GetSession(r, "username")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if value != "testuser" {
		t.Errorf("Expected session value 'testuser', got '%v'", value)
	}
}

func TestClearSession(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	session, _ := sessionStore.Get(r, "session-name")
	session.Values["username"] = "testuser"
	err := session.Save(r, w)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	err = ClearSession(w, r)
	if err != nil {
		t.Fatalf("ClearSession failed: %v", err)
	}

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Del("Cookie")

	session, err = sessionStore.Get(r, "session-name")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.Values["username"] != nil {
		t.Errorf("Expected session value to be cleared, but got '%v'", session.Values["username"])
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup a router with the AuthMiddleware
	router := gin.Default()
	router.Use(SessionMiddleware())

	protected := router.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to the dashboard"})
		})
	}

	// Test case 1: Access without session (should fail)
	t.Run("Unauthorized Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/dashboard", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized", response["error"])
	})

	// Test case 2: Access with valid session (should succeed)
	t.Run("Authorized Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/dashboard", nil)

		// Simulate a logged-in user by setting a valid session
		session := sessions.NewSession(sessionStore, "session-name")
		session.Values["user_id"] = 1
		session.Save(req, w)

		req.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Welcome to the dashboard", response["message"])
	})
}
