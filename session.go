package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"gopkg.in/yaml.v2"
)

type Config struct {
	SessionSecretKey string `yaml:"session_secret_key"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

var sessionStore *sessions.CookieStore

func init() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if config.SessionSecretKey == "" {
		log.Fatal("Session secret key is not set in the configuration file")
	}

	sessionStore = sessions.NewCookieStore([]byte(config.SessionSecretKey))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
		Secure:   false,
	}
}

func SetSession(w http.ResponseWriter, r *http.Request, name string, value interface{}) error {
	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		return err
	}
	session.Values[name] = value
	return session.Save(r, w)
}

func GetSession(r *http.Request, name string) (interface{}, error) {
	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		return nil, err
	}
	return session.Values[name], nil
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		return err
	}

	// Explicitly delete all session values
	for key := range session.Values {
		delete(session.Values, key)
	}

	// Invalidate the session cookie
	session.Options.MaxAge = -1

	// Save the session to apply the changes
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}
