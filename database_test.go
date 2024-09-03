package main

import (
	"database/sql"
	"os"
	"testing"
)

func openTestDB(t *testing.T) *sql.DB {
	os.Remove("users_test.db")

	db, err := sql.Open("sqlite", "./users_test.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestReadUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "readtestuser"

	userID, err := CreateUserIfNotExists(db, username)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}
}

func TestUpdateUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "updatetestuser"

	userID, err := CreateUserIfNotExists(db, username)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	updatedUsername := "updateduser"

	err = UpdateUser(db, int(userID), updatedUsername)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.Username != updatedUsername {
		t.Errorf("Expected username %s, got %s", updatedUsername, user.Username)
	}
}

func TestDeleteUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "deletetestuser"

	userID, err := CreateUserIfNotExists(db, username)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	err = DeleteUser(db, int(userID))
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err == nil {
		t.Errorf("Expected error when reading deleted user, got none. User: %+v", user)
	}
}

func TestCreateUserIfNotExists(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "uniqueuser"

	userID, err := CreateUserIfNotExists(db, username)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed on first attempt: %v", err)
	}

	if userID <= 0 {
		t.Errorf("Expected valid user ID, got %d", userID)
	}

	_, err = CreateUserIfNotExists(db, username)
	if err == nil {
		t.Fatalf("Expected error for duplicate user creation, but got none")
	} else {
		expectedError := "user with this username already exists"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
		}
	}
}

func TestValidateUsername(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	testCases := []struct {
		username string
		isValid  bool
	}{
		{"validuser", true},
		{"user123", true},
		{"user_name", true},
		{"user_name_", true},
		{"username1234", true},
		{"user", true},
		{"1username", false}, // starts with a digit
		{"user-name", false}, // contains a hyphen
		{"UserName", false},  // contains uppercase letters
		{"user name", false}, // contains a space
		{"us", false},        // too short
		{"thisisaverylongusernameover24characters", false}, // too long
		{"user@name", false}, // contains an invalid character
	}

	for _, tc := range testCases {
		err := validateUsername(tc.username)
		if (err == nil) != tc.isValid {
			t.Errorf("Expected validity of username '%s' to be %v, got error: %v", tc.username, tc.isValid, err)
		}
	}
}
