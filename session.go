package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

	for key := range session.Values {
		delete(session.Values, key)
	}

	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := sessionStore.Get(c.Request, "session-name")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
			return
		}

		c.Set("session", session)

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attempt to retrieve the session
		session, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Session not found"})
			c.Abort()
			return
		}

		userID, ok := session.(*sessions.Session).Values["user_id"]
		if !ok || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found in session"})
			c.Abort()
			return
		}

		// User is authenticated, continue to the next handler
		c.Next()
	}
}
