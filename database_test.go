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
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL
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
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
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

	if !CheckPasswordHash(password, user.PasswordHash) {
		t.Errorf("Password hash did not match original password after username update")
	}
}
func TestUpdateUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "updatetestuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	updatedUsername := "updateduser"
	err = UpdateUser(db, int(userID), updatedUsername, "")
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

	if !CheckPasswordHash(password, user.PasswordHash) {
		t.Errorf("Password hash did not match original password after username update")
	}
}

func TestDeleteUser(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "deletetestuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
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
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed on first attempt: %v", err)
	}

	if userID <= 0 {
		t.Errorf("Expected valid user ID, got %d", userID)
	}

	_, err = CreateUserIfNotExists(db, username, password)
	if err == nil {
		t.Fatalf("Expected error for duplicate user creation, but got none")
	} else {
		expectedError := "user with this username already exists"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
		}
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	password := "ValidP@ssw0rd"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPasswordHash(password, hashedPassword) {
		t.Errorf("Password did not match hashed password")
	}

	if CheckPasswordHash("wrongpassword", hashedPassword) {
		t.Errorf("Expected hash check to fail with wrong password")
	}
}

func TestCreateUserWithPassword(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "secureuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.PasswordHash == "" {
		t.Fatalf("Password hash was not stored correctly.")
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		t.Errorf("Stored password hash did not match the original password")
	}
}

func TestUpdateUsernameOnly(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "initialuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	newUsername := "updateduser"
	err = UpdateUser(db, int(userID), newUsername, "")
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.Username != newUsername {
		t.Errorf("Expected username %s, got %s", newUsername, user.Username)
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		t.Errorf("Password hash did not match original password after username update")
	}
}

func TestUpdatePasswordOnly(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "initialuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	newPassword := "UpdatedP@ssw0rd"
	err = UpdateUser(db, int(userID), "", newPassword)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}

	if !CheckPasswordHash(newPassword, user.PasswordHash) {
		t.Errorf("Stored password hash did not match updated password")
	}
}

func TestUpdateUsernameAndPassword(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "initialuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	newUsername := "updateduser"
	newPassword := "UpdatedP@ssw0rd"
	err = UpdateUser(db, int(userID), newUsername, newPassword)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.Username != newUsername {
		t.Errorf("Expected username %s, got %s", newUsername, user.Username)
	}

	if !CheckPasswordHash(newPassword, user.PasswordHash) {
		t.Errorf("Stored password hash did not match updated password")
	}
}

func TestUpdateNoFields(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	username := "initialuser"
	password := "ValidP@ssw0rd"

	userID, err := CreateUserIfNotExists(db, username, password)
	if err != nil {
		t.Fatalf("CreateUserIfNotExists failed: %v", err)
	}

	err = UpdateUser(db, int(userID), "", "")
	if err == nil {
		t.Fatalf("Expected error when updating with no fields, but got none")
	}

	user, err := ReadUser(db, int(userID))
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}

	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		t.Errorf("Stored password hash did not match original password")
	}
}
