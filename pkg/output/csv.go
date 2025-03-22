package output

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mikev/spotify-analysis/pkg/processor"
)

// CSVWriter handles writing track data to CSV files
type CSVWriter struct {
	outputDir string
	overwrite bool
}

// NewCSVWriter creates a new CSV writer
func NewCSVWriter(outputDir string, overwrite bool) (*CSVWriter, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	return &CSVWriter{
		outputDir: outputDir,
		overwrite: overwrite,
	}, nil
}

// WriteTracks writes track data to CSV files
func (w *CSVWriter) WriteTracks(tracks map[string][]processor.TrackData) error {
	// Write user tracks
	if err := w.writeToCSV("user_playlists.csv", tracks["user"]); err != nil {
		return err
	}

	// Write other tracks if they exist
	if len(tracks["other"]) > 0 {
		if err := w.writeToCSV("other_playlists.csv", tracks["other"]); err != nil {
			return err
		}
	}

	return nil
}

// writeToCSV writes track data to a specific CSV file
func (w *CSVWriter) writeToCSV(filename string, tracks []processor.TrackData) error {
	filepath := filepath.Join(w.outputDir, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); err == nil {
		if !w.overwrite {
			return fmt.Errorf("file %s already exists and overwrite is disabled", filename)
		}
		log.Printf("File %s already exists. Removing it...", filename)
		if err := os.Remove(filepath); err != nil {
			return fmt.Errorf("failed to remove existing file %s: %v", filename, err)
		}
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer file.Close()

	// Add UTF-8 BOM for proper Excel encoding
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := []string{"Playlist", "Track Name", "Artist(s)", "Album", "Release Date", "Release Year", "NotInTopTrackPlaylist"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	// Write track data with progress logging
	totalTracks := len(tracks)
	log.Printf("Writing %d tracks to %s...", totalTracks, filename)

	for i, track := range tracks {
		row := []string{
			track.PlaylistName,
			track.TrackName,
			track.Artists,
			track.Album,
			track.ReleaseDate,
			track.ReleaseYear,
			track.NotInTopTracks,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write track %s: %v", track.TrackName, err)
		}

		// Log progress every 100 tracks
		if (i+1)%100 == 0 {
			log.Printf("Progress: %d/%d tracks written to %s", i+1, totalTracks, filename)
		}
	}

	log.Printf("Successfully wrote %d tracks to %s", totalTracks, filename)
	return nil
}
