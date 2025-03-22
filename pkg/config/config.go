package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	ClientID              string
	ClientSecret          string
	RedirectURI           string
	Port                  int
	TopTracksPattern      string
	StartYear             string
	EndYear               string
	IncludeOtherPlaylists bool
	OverwriteFiles        bool
	LogFile               string
	LogRotateSize         string
	LogKeepFiles          int
}

// LoadConfig loads and validates all configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Get required environment variables
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	port := os.Getenv("SPOTIFY_PORT")
	topTracksPattern := os.Getenv("SPOTIFY_TOP_TRACKS_PATTERN")
	startYear := os.Getenv("SPOTIFY_START_YEAR")
	endYear := os.Getenv("SPOTIFY_END_YEAR")
	includeOtherPlaylists := os.Getenv("SPOTIFY_INCLUDE_OTHER_PLAYLISTS")
	overwriteFiles := os.Getenv("SPOTIFY_OVERWRITE_FILES")
	logFile := os.Getenv("SPOTIFY_LOG_FILE")
	logRotateSize := os.Getenv("SPOTIFY_LOG_ROTATE_SIZE")
	logKeepFiles := os.Getenv("SPOTIFY_LOG_KEEP_FILES")

	// Log configuration values (excluding sensitive data)
	log.Printf("Configuration loaded:")
	log.Printf("  Redirect URI: %s", redirectURI)
	log.Printf("  Port: %s", port)
	log.Printf("  Top Tracks Pattern: %s", topTracksPattern)
	log.Printf("  Start Year: %s", startYear)
	log.Printf("  End Year: %s", endYear)
	log.Printf("  Include Other Playlists: %s", includeOtherPlaylists)
	log.Printf("  Overwrite Files: %s", overwriteFiles)
	log.Printf("  Log File: %s", logFile)
	log.Printf("  Log Rotate Size: %s", logRotateSize)
	log.Printf("  Log Keep Files: %s", logKeepFiles)

	// Validate required variables
	if clientID == "" || clientSecret == "" || redirectURI == "" || port == "" ||
		topTracksPattern == "" || startYear == "" || endYear == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	// Convert port to integer
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	// Parse overwrite files setting with explicit logging
	var overwriteFilesBool bool
	switch strings.ToLower(overwriteFiles) {
	case "true":
		overwriteFilesBool = true
		log.Println("File overwriting enabled")
	case "false":
		overwriteFilesBool = false
		log.Println("File overwriting disabled")
	default:
		overwriteFilesBool = true
		log.Printf("SPOTIFY_OVERWRITE_FILES not set or invalid (%s), defaulting to true", overwriteFiles)
	}

	// Set default logging values if not specified
	if logFile == "" {
		logFile = "logs/spotify-analysis.log"
		log.Println("Using default log file path")
	}
	if logRotateSize == "" {
		logRotateSize = "10MB"
		log.Println("Using default log rotate size")
	}
	if logKeepFiles == "" {
		logKeepFiles = "7"
		log.Println("Using default log keep files count")
	}

	// Convert log keep files to integer
	keepFiles, err := strconv.Atoi(logKeepFiles)
	if err != nil {
		return nil, fmt.Errorf("invalid log keep files value: %v", err)
	}

	return &Config{
		ClientID:              clientID,
		ClientSecret:          clientSecret,
		RedirectURI:           redirectURI,
		Port:                  portNum,
		TopTracksPattern:      topTracksPattern,
		StartYear:             startYear,
		EndYear:               endYear,
		IncludeOtherPlaylists: includeOtherPlaylists == "true",
		OverwriteFiles:        overwriteFilesBool,
		LogFile:               logFile,
		LogRotateSize:         logRotateSize,
		LogKeepFiles:          keepFiles,
	}, nil
}
