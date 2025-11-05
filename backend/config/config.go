package config

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Config holds application configuration
type Config struct {
	DatabaseURL   string
	ServerAddress string
	CacheTimeout  int // milliseconds
}

// Load reads configuration from .env or environment variables
func Load() *Config {
	// Load .env file (ignore if missing)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set in environment or .env file")
	}

	serverAddr := os.Getenv("SERVER_ADDRESS")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	cacheTimeoutStr := os.Getenv("CACHE_TIMEOUT")
	cacheTimeout, err := strconv.Atoi(cacheTimeoutStr)
	if err != nil || cacheTimeout == 0 {
		cacheTimeout = 100
	}

	return &Config{
		DatabaseURL:   dbURL,
		ServerAddress: serverAddr,
		CacheTimeout:  cacheTimeout,
	}
}

// InitDB initializes database connection
func InitDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}
