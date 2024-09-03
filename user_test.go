package main

import (
	"net/http/httptest"
	"testing"
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
