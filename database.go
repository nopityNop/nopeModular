package main

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/argon2"
	_ "modernc.org/sqlite"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

const (
	ArgonTime    = 1
	ArgonMemory  = 64 * 1024 // 64 MB
	ArgonThreads = 4
	ArgonKeyLen  = 32
	ArgonSaltLen = 16
)

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./users.db")
	if err != nil {
		return nil, err
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
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

func CreateUser(db *sql.DB, username, password string) (int64, error) {
	if err := validateUsername(username); err != nil {
		return 0, err
	}

	if err := validatePassword(password); err != nil {
		return 0, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return 0, err
	}

	result, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hashedPassword)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func CreateUserIfNotExists(db *sql.DB, username, password string) (int64, error) {
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

	return CreateUser(db, username, password)

}

func ReadUser(db *sql.DB, id int) (*User, error) {
	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE id = ?", id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUser(db *sql.DB, id int, username, password string) error {
	if username == "" && password == "" {
		return errors.New("at least one of username or password must be provided")
	}

	var err error
	updateFields := make([]string, 0)
	updateArgs := make([]interface{}, 0)

	if username != "" {
		if err = validateUsername(username); err != nil {
			return err
		}
		updateFields = append(updateFields, "username = ?")
		updateArgs = append(updateArgs, username)
	}

	if password != "" {
		hashedPassword, err := HashPassword(password)
		if err != nil {
			return err
		}
		updateFields = append(updateFields, "password_hash = ?")
		updateArgs = append(updateArgs, hashedPassword)
	}

	updateArgs = append(updateArgs, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updateFields, ", "))

	_, err = db.Exec(query, updateArgs...)
	return err
}

func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, ArgonSaltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func HashPassword(password string) (string, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	encodedHash := fmt.Sprintf("%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
	return encodedHash, nil
}

func CheckPasswordHash(password, encodedHash string) bool {
	parts := split(encodedHash, '$')
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	return subtle.ConstantTimeCompare(hash, computedHash) == 1
}

func split(s string, delim byte) []string {
	var result []string
	var start int
	for i := 0; i < len(s); i++ {
		if s[i] == delim {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", username)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func EnsureTestUser(db *sql.DB) error {
	username := "test"
	password := "Test@1234"

	exists, err := UserExists(db, username)
	if err != nil {
		return err
	}

	if !exists {
		_, err := CreateUser(db, username, password)
		if err != nil {
			return err
		}
		log.Printf("Created test user with username: %s", username)
	} else {
		log.Printf("Test user already exists: %s", username)
	}

	return nil
}
