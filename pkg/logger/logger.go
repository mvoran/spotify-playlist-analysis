package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds logging configuration
type Config struct {
	LogFile    string
	RotateSize string
	KeepFiles  int
}

// InitLogger initializes the logger with file output
func InitLogger(cfg *Config) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Set up multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Configure logger
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return nil
}

// RotateLog rotates the log file if it exceeds the size limit
func RotateLog(cfg *Config) error {
	// Check if file exists and get its size
	fileInfo, err := os.Stat(cfg.LogFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, no need to rotate
		}
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	// Parse rotate size (e.g., "10MB")
	size, err := parseSize(cfg.RotateSize)
	if err != nil {
		return fmt.Errorf("invalid rotate size: %v", err)
	}

	// If file is smaller than rotate size, no need to rotate
	if fileInfo.Size() < size {
		return nil
	}

	// Generate new filename with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	dir := filepath.Dir(cfg.LogFile)
	base := filepath.Base(cfg.LogFile)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	newFile := filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, timestamp, ext))

	// Rename current log file
	if err := os.Rename(cfg.LogFile, newFile); err != nil {
		return fmt.Errorf("failed to rotate log file: %v", err)
	}

	// Clean up old log files
	if err := cleanupOldLogs(dir, name, cfg.KeepFiles); err != nil {
		log.Printf("Warning: failed to cleanup old log files: %v", err)
	}

	return nil
}

// parseSize converts a size string (e.g., "10MB") to bytes
func parseSize(sizeStr string) (int64, error) {
	var size int64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%d%s", &size, &unit)
	if err != nil {
		return 0, err
	}

	switch strings.ToUpper(unit) {
	case "B":
		return size, nil
	case "KB":
		return size * 1024, nil
	case "MB":
		return size * 1024 * 1024, nil
	case "GB":
		return size * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}
}

// cleanupOldLogs removes old log files, keeping only the specified number
func cleanupOldLogs(dir, name string, keepFiles int) error {
	pattern := filepath.Join(dir, name+"-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Sort matches by name (which includes timestamp) in descending order
	// This ensures we keep the most recent files
	sortFilesByTime(matches)

	// Remove excess files
	for i := keepFiles; i < len(matches); i++ {
		if err := os.Remove(matches[i]); err != nil {
			log.Printf("Warning: failed to remove old log file %s: %v", matches[i], err)
		}
	}

	return nil
}

// sortFilesByTime sorts file paths by their embedded timestamps
func sortFilesByTime(files []string) {
	// Simple bubble sort (fine for small number of files)
	for i := 0; i < len(files)-1; i++ {
		for j := 0; j < len(files)-i-1; j++ {
			if files[j] < files[j+1] {
				files[j], files[j+1] = files[j+1], files[j]
			}
		}
	}
}
