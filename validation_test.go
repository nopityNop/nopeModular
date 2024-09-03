package main

import (
	"testing"
)

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
		{"1username", false},
		{"user-name", false},
		{"UserName", false},
		{"user name", false},
		{"us", false},
		{"thisisaverylongusernameover24characters", false},
		{"user@name", false},
	}

	mockPassword := "ValidP@ssw0rd"

	for _, tc := range testCases {
		_, err := CreateUserIfNotExists(db, tc.username, mockPassword)
		if (err == nil) != tc.isValid {
			t.Errorf("Expected validity of username '%s' to be %v, got error: %v", tc.username, tc.isValid, err)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		password string
		isValid  bool
	}{
		{"P@ssw0rd", true},
		{"password", false},
		{"PASSWORD", false},
		{"Passw0rd", false},
		{"P@ssword", false},
		{"P@ss1", false},
		{"ValidP@ssword123", true},
	}

	for _, tc := range testCases {
		err := validatePassword(tc.password)
		if (err == nil) != tc.isValid {
			t.Errorf("Expected validity of password '%s' to be %v, got error: %v", tc.password, tc.isValid, err)
		}
	}
}
