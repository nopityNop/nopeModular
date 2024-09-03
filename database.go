package main

import (
	"database/sql"
	"errors"
	"regexp"

	_ "modernc.org/sqlite"
)

type User struct {
	ID       int
	Username string
}

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./users.db")
	if err != nil {
		return nil, err
	}

	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
}

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

func UserExists(db *sql.DB, username string) (bool, error) {
	var exists bool
	query := "SELECT COUNT(1) FROM users WHERE username = ?"
	err := db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateUser(db *sql.DB, username string) (int64, error) {
	if err := validateUsername(username); err != nil {
		return 0, err
	}

	result, err := db.Exec("INSERT INTO users (username) VALUES (?)", username)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func CreateUserIfNotExists(db *sql.DB, username string) (int64, error) {
	if err := validateUsername(username); err != nil {
		return 0, err
	}

	exists, err := UserExists(db, username)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, errors.New("user with this username already exists")
	}

	return CreateUser(db, username)
}

func ReadUser(db *sql.DB, id int) (*User, error) {
	row := db.QueryRow("SELECT id, username FROM users WHERE id = ?", id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUser(db *sql.DB, id int, username string) error {
	if err := validateUsername(username); err != nil {
		return err
	}

	_, err := db.Exec("UPDATE users SET username = ? WHERE id = ?", username, id)
	return err
}

func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
