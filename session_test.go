package main

import (
	"net/http/httptest"
	"testing"
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
