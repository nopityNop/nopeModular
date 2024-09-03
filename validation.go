package main

import (
	"errors"
	"regexp"

	_ "modernc.org/sqlite"
)

func validateUsername(username string) error {
	if len(username) < 4 || len(username) > 24 {
		return errors.New("username must be between 4 and 24 characters long")
	}

	matched, err := regexp.MatchString(`^[a-z][a-z0-9_]*$`, username)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("username must start with a letter and can only contain lowercase letters, digits, and underscores")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	specialCharPattern := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)

	for _, char := range password {
		switch {
		case 'a' <= char && char <= 'z':
			hasLower = true
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case specialCharPattern.MatchString(string(char)):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
