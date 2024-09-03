package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

func TestLoginUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "loginuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/login", nil)

	loggedInUserID, err := LoginUser(w, r, db, username, password)
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}

	if loggedInUserID != int(userID) {
		t.Errorf("Expected user ID %d, got %d", userID, loggedInUserID)
	}

	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.Values["user_id"] != int(userID) {
		t.Errorf("Expected session user_id to be %d, but got %v", userID, session.Values["user_id"])
	}
}

func TestLoginUserWithSession(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "loginuser"
	password := "ValidP@ssw0rd"

	_, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/login", nil)

	_, err = LoginUser(w, r, db, username, password)
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}

	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.Values["user_id"] == nil {
		t.Errorf("Expected session to contain user_id, but it was nil")
	}
}

func TestLogoutUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "logoutuser"
	password := "ValidP@ssw0rd"

	_, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/login", nil)

	_, err = LoginUser(w, r, db, username, password)
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}

	err = LogoutUser(w, r)
	if err != nil {
		t.Fatalf("LogoutUser failed: %v", err)
	}

	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.Values["user_id"] != nil {
		t.Errorf("Expected session to be cleared, but got '%v'", session.Values["user_id"])
	}
}

func TestLogoutHandler(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "testuser"
	password := "ValidP@ssw0rd"
	_, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/login", nil)
	r.PostForm = url.Values{}
	r.PostForm.Set("username", username)
	r.PostForm.Set("password", password)

	router := gin.Default()
	router.POST("/login", func(c *gin.Context) {
		LoginHandler(c, OpenDB)
	})
	router.ServeHTTP(w, r)

	r = httptest.NewRequest("GET", "/logout", nil)
	w = httptest.NewRecorder()
	router.GET("/logout", LogoutHandler)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["message"] != "Logout successful" {
		t.Errorf("Expected message 'Logout successful', got '%v'", response["message"])
	}
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup a shared database connection for the test
	db := openTestDB(t)
	defer db.Close()

	// Create a mock user for testing purposes
	username := "testuser"
	password := "ValidP@ssw0rd"

	_, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}
	t.Logf("Created user with ID: 1")

	// Define a function that returns the shared DB connection
	dbFunc := func() (*sql.DB, error) {
		return db, nil
	}

	// Test Case 1: Successful login
	t.Run("Successful Login", func(t *testing.T) {
		router := gin.Default() // Create a new router for each test case
		router.POST("/login", func(c *gin.Context) {
			LoginHandler(c, dbFunc) // Use the shared DB connection
		})

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("username", username)
		form.Add("password", password)
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = form

		router.ServeHTTP(w, req)

		// Debugging output
		if w.Code != http.StatusOK {
			t.Logf("Received status: %d, Body: %s", w.Code, w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Login successful", response["message"])
		assert.NotNil(t, response["user_id"])
	})

	// Test Case 2: Invalid login credentials
	t.Run("Invalid Credentials", func(t *testing.T) {
		router := gin.Default() // Create a new router for each test case
		router.POST("/login", func(c *gin.Context) {
			LoginHandler(c, dbFunc) // Use the shared DB connection
		})

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("username", "wronguser")
		form.Add("password", "wrongpassword")
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = form

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid username or password", response["error"])
	})

	// Test Case 3: Database connection failure
	t.Run("Database Connection Failure", func(t *testing.T) {
		// Define a mock function that simulates a database connection failure
		mockDBFunc := func() (*sql.DB, error) {
			return nil, errors.New("simulated connection failure")
		}

		router := gin.Default() // Create a new router for each test case
		router.POST("/login", func(c *gin.Context) {
			LoginHandler(c, mockDBFunc)
		})

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("username", username)
		form.Add("password", password)
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = form

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to connect to database", response["error"])
	})
}

func TestDashboardHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup a router with the AuthMiddleware and DashboardHandler
	router := gin.Default()
	router.Use(SessionMiddleware())

	protected := router.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/dashboard", DashboardHandler)
	}

	// Test case: Access with valid session (should succeed)
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
		assert.Equal(t, "Welcome to the dashboard, User 1!", response["message"])
		assert.Equal(t, 1, int(response["user_id"].(float64)))
	})

	// Test case: Access without session (should fail)
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
}
