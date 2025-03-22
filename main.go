package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikev/spotify-analysis/pkg/config"
	"github.com/mikev/spotify-analysis/pkg/logger"
	"github.com/mikev/spotify-analysis/pkg/output"
	"github.com/mikev/spotify-analysis/pkg/processor"
	"github.com/mikev/spotify-analysis/pkg/spotify"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logCfg := &logger.Config{
		LogFile:    cfg.LogFile,
		RotateSize: cfg.LogRotateSize,
		KeepFiles:  cfg.LogKeepFiles,
	}
	if err := logger.InitLogger(logCfg); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Spotify client
	client, err := spotify.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Spotify client: %v", err)
	}
	// Ensure client cleanup on exit
	defer client.Cleanup()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal. Cleaning up...")
		client.Cleanup()
		os.Exit(0)
	}()

	// Initialize playlist processor
	processor, err := processor.NewPlaylistProcessor(client.Client, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize playlist processor: %v", err)
	}

	// Process playlists
	tracks, err := processor.ProcessPlaylists()
	if err != nil {
		log.Fatalf("Failed to process playlists: %v", err)
	}

	// Initialize CSV writer
	writer, err := output.NewCSVWriter("playlists", cfg.OverwriteFiles)
	if err != nil {
		log.Fatalf("Failed to initialize CSV writer: %v", err)
	}

	// Write tracks to CSV files
	if err := writer.WriteTracks(tracks); err != nil {
		log.Fatalf("Failed to write tracks to CSV: %v", err)
	}

	fmt.Println("All playlists have been processed!")
}
