package processor

import (
	"fmt"
	"log"
	"strings"

	"github.com/mikev/spotify-analysis/pkg/config"
	"github.com/zmb3/spotify"
)

// TrackInfo holds the unique identifier for a track
type TrackInfo struct {
	ID   string
	Name string
}

// PlaylistProcessor handles playlist and track processing
type PlaylistProcessor struct {
	client       *spotify.Client
	cfg          *config.Config
	topTracksMap map[string]TrackInfo
	userID       string
}

// NewPlaylistProcessor creates a new playlist processor
func NewPlaylistProcessor(client *spotify.Client, cfg *config.Config) (*PlaylistProcessor, error) {
	user, err := client.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %v", err)
	}

	return &PlaylistProcessor{
		client:       client,
		cfg:          cfg,
		topTracksMap: make(map[string]TrackInfo),
		userID:       user.ID,
	}, nil
}

// ProcessPlaylists processes all playlists and returns track data
func (p *PlaylistProcessor) ProcessPlaylists() (map[string][]TrackData, error) {
	log.Println("Starting playlist processing...")

	// First, collect all tracks from top tracks playlists
	log.Println("Collecting tracks from top tracks playlists...")
	if err := p.collectTopTracks(); err != nil {
		return nil, err
	}

	// Get all playlists
	log.Println("Fetching all playlists...")
	allPlaylists, err := p.getAllPlaylists()
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d total playlists to process", len(allPlaylists))

	// Process playlists and collect track data
	userTracks := make([]TrackData, 0)
	otherTracks := make([]TrackData, 0)

	for i, playlist := range allPlaylists {
		log.Printf("Processing playlist %d/%d: %s", i+1, len(allPlaylists), playlist.Name)
		tracks, err := p.processPlaylist(playlist)
		if err != nil {
			log.Printf("Error processing playlist %s: %v", playlist.Name, err)
			continue
		}

		if playlist.Owner.ID != p.userID {
			if p.cfg.IncludeOtherPlaylists {
				log.Printf("Adding %d tracks from other user's playlist: %s", len(tracks), playlist.Name)
				otherTracks = append(otherTracks, tracks...)
			}
		} else {
			log.Printf("Adding %d tracks from your playlist: %s", len(tracks), playlist.Name)
			userTracks = append(userTracks, tracks...)
		}
	}

	log.Printf("Processing complete. Found %d tracks in your playlists and %d tracks in other playlists",
		len(userTracks), len(otherTracks))

	return map[string][]TrackData{
		"user":  userTracks,
		"other": otherTracks,
	}, nil
}

// collectTopTracks collects all tracks from top tracks playlists
func (p *PlaylistProcessor) collectTopTracks() error {
	playlists, err := p.getAllPlaylists()
	if err != nil {
		return err
	}

	for _, playlist := range playlists {
		normalizedName := normalizeQuotes(strings.ToLower(playlist.Name))
		if strings.Contains(normalizedName, strings.ToLower(p.cfg.TopTracksPattern)) {
			fmt.Printf("Processing top tracks playlist: %s\n", playlist.Name)
			if err := p.processPlaylistTracks(playlist.ID, func(track spotify.FullTrack) {
				p.topTracksMap[string(track.ID)] = TrackInfo{
					ID:   string(track.ID),
					Name: track.Name,
				}
			}); err != nil {
				return err
			}
		}
	}

	fmt.Printf("Found %d unique tracks in top tracks playlists\n", len(p.topTracksMap))
	return nil
}

// getAllPlaylists retrieves all playlists with pagination
func (p *PlaylistProcessor) getAllPlaylists() ([]spotify.SimplePlaylist, error) {
	var allPlaylists []spotify.SimplePlaylist
	offset := 0
	limit := 50 // Maximum allowed by Spotify API

	for {
		playlists, err := p.client.CurrentUsersPlaylistsOpt(&spotify.Options{
			Limit:  &limit,
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get playlists: %v", err)
		}

		allPlaylists = append(allPlaylists, playlists.Playlists...)

		if len(playlists.Playlists) < limit {
			break
		}

		offset += limit
	}

	return allPlaylists, nil
}

// processPlaylist processes a single playlist and returns its track data
func (p *PlaylistProcessor) processPlaylist(playlist spotify.SimplePlaylist) ([]TrackData, error) {
	log.Printf("Starting to process playlist: %s", playlist.Name)
	var tracks []TrackData
	err := p.processPlaylistTracks(playlist.ID, func(track spotify.FullTrack) {
		tracks = append(tracks, p.createTrackData(playlist.Name, track))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process playlist %s: %v", playlist.Name, err)
	}
	log.Printf("Finished processing playlist %s: found %d tracks", playlist.Name, len(tracks))
	return tracks, nil
}

// processPlaylistTracks processes all tracks in a playlist with pagination
func (p *PlaylistProcessor) processPlaylistTracks(playlistID spotify.ID, processTrack func(spotify.FullTrack)) error {
	offset := 0
	limit := 100 // Maximum allowed by Spotify API
	totalProcessed := 0

	for {
		log.Printf("Fetching tracks from playlist (offset: %d, limit: %d)...", offset, limit)
		tracks, err := p.client.GetPlaylistTracksOpt(playlistID, &spotify.Options{
			Limit:  &limit,
			Offset: &offset,
		}, "")
		if err != nil {
			return fmt.Errorf("failed to get tracks: %v", err)
		}

		log.Printf("Processing %d tracks from current batch...", len(tracks.Tracks))
		for _, item := range tracks.Tracks {
			processTrack(item.Track)
			totalProcessed++
		}

		if len(tracks.Tracks) < limit {
			break
		}

		offset += limit
	}

	log.Printf("Finished processing all tracks from playlist. Total tracks processed: %d", totalProcessed)
	return nil
}

// TrackData represents processed track information
type TrackData struct {
	PlaylistName   string
	TrackName      string
	Artists        string
	Album          string
	ReleaseDate    string
	ReleaseYear    string
	NotInTopTracks string
}

// createTrackData creates a TrackData object from a Spotify track
func (p *PlaylistProcessor) createTrackData(playlistName string, track spotify.FullTrack) TrackData {
	artists := ""
	for i, artist := range track.Artists {
		if i > 0 {
			artists += ", "
		}
		artists += artist.Name
	}

	releaseYear := ""
	if track.Album.ReleaseDate != "" {
		parts := strings.Split(track.Album.ReleaseDate, "-")
		if len(parts) > 0 {
			releaseYear = parts[0]
		}
	}

	notInTopTracks := ""
	if releaseYear != "" {
		if releaseYear >= p.cfg.StartYear && releaseYear <= p.cfg.EndYear {
			trackID := string(track.ID)
			if _, exists := p.topTracksMap[trackID]; !exists {
				notInTopTracks = "TRUE"
			}
		}
	}

	return TrackData{
		PlaylistName:   playlistName,
		TrackName:      track.Name,
		Artists:        artists,
		Album:          track.Album.Name,
		ReleaseDate:    track.Album.ReleaseDate,
		ReleaseYear:    releaseYear,
		NotInTopTracks: notInTopTracks,
	}
}

// normalizeQuotes replaces smart quotes with regular quotes
func normalizeQuotes(s string) string {
	s = strings.ReplaceAll(s, "\u2019", "'")  // Replace right single quotation mark
	s = strings.ReplaceAll(s, "\u2018", "'")  // Replace left single quotation mark
	s = strings.ReplaceAll(s, "\u201C", "\"") // Replace left double quotation mark
	s = strings.ReplaceAll(s, "\u201D", "\"") // Replace right double quotation mark
	return s
}
