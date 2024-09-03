package main

import (
	"database/sql"
	"errors"
	"net/http"

	_ "modernc.org/sqlite"
)

func LoginUser(w http.ResponseWriter, r *http.Request, db *sql.DB, username, password string) (int, error) {
	user, err := GetUserByUsername(db, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("invalid username or password")
		}
		return 0, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return 0, errors.New("invalid username or password")
	}

	err = SetSession(w, r, "user_id", user.ID)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func LogoutUser(w http.ResponseWriter, r *http.Request) error {
	return ClearSession(w, r)
}
