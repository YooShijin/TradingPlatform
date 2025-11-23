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
	DatabaseURL     string
	ServerAddress   string
	CacheTimeout    int // time in miliseconds for order_handler
	CacheSize       int
	CacheTTLSeconds int // time in seconds for get or put (recent trades)

}

// We load and store configuration from our .env file (environment variable)
func Load() *Config {

	// Load .env file (ignore if missing)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in .env")
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
	cacheSizeStr := os.Getenv("CACHESIZE")
	cacheSize, err := strconv.Atoi((cacheSizeStr))
	if err != nil || cacheSize == 0 {
		cacheSize = 10
	}
	cacheTTLSecondsStr := os.Getenv("CACHETTLSECONDS")
	cacheTTLSeconds, err := strconv.Atoi((cacheTTLSecondsStr))
	if err != nil || cacheTTLSeconds == 0 {
		cacheTTLSeconds = 10
	}

	return &Config{
		DatabaseURL:     dbURL,
		ServerAddress:   serverAddr,
		CacheTimeout:    cacheTimeout,
		CacheSize:       cacheSize,
		CacheTTLSeconds: cacheTTLSeconds,
	}
}

// InitDB function will initialize the database connection (capital letteer )
func InitDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)

	return db, nil
}
