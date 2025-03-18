package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify"
)

var (
	auth  spotify.Authenticator
	ch    = make(chan *spotify.Client)
	state = "some-random-state-key"
)

// TrackInfo holds the unique identifier for a track
type TrackInfo struct {
	ID   string
	Name string
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get Spotify credentials from environment variables
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	port := os.Getenv("SPOTIFY_PORT")
	topTracksPattern := os.Getenv("SPOTIFY_TOP_TRACKS_PATTERN")
	startYear := os.Getenv("SPOTIFY_START_YEAR")
	endYear := os.Getenv("SPOTIFY_END_YEAR")
	includeOtherPlaylists := os.Getenv("SPOTIFY_INCLUDE_OTHER_PLAYLISTS")

	if clientID == "" || clientSecret == "" || redirectURI == "" || port == "" || topTracksPattern == "" || startYear == "" || endYear == "" {
		log.Fatal("SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, SPOTIFY_REDIRECT_URI, SPOTIFY_PORT, SPOTIFY_TOP_TRACKS_PATTERN, SPOTIFY_START_YEAR, and SPOTIFY_END_YEAR must be set in .env file")
	}

	// Convert port to integer
	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	// Initialize the authenticator with the redirect URI
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadCollaborative)
	auth.SetAuthInfo(clientID, clientSecret)

	// Start local server to receive the callback
	http.HandleFunc("/callback", completeAuth)

	// Start server in a goroutine
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", portNum), nil); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Open the browser for authentication
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	openBrowser(url)

	// Wait for auth to complete
	client := <-ch

	// Get user's playlists
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatalf("failed to get current user: %v", err)
	}

	fmt.Printf("Fetching playlists for user: %s\n", user.DisplayName)

	// Get all playlists with pagination
	var allPlaylists []spotify.SimplePlaylist
	offset := 0
	limit := 50 // Maximum allowed by Spotify API

	for {
		playlists, err := client.CurrentUsersPlaylistsOpt(&spotify.Options{
			Limit:  &limit,
			Offset: &offset,
		})
		if err != nil {
			log.Fatalf("failed to get playlists: %v", err)
		}

		allPlaylists = append(allPlaylists, playlists.Playlists...)

		// Check if we've reached the end of the playlists
		if len(playlists.Playlists) < limit {
			break
		}

		offset += limit
	}

	fmt.Printf("Found %d total playlists\n", len(allPlaylists))

	// First, collect all tracks from top tracks playlists
	topTracksMap := make(map[string]TrackInfo)
	for _, playlist := range allPlaylists {
		normalizedName := normalizeQuotes(strings.ToLower(playlist.Name))
		if strings.Contains(normalizedName, strings.ToLower(topTracksPattern)) {
			fmt.Printf("Processing top tracks playlist: %s\n", playlist.Name)
			// Process all tracks in the playlist with pagination
			offset := 0
			limit := 100 // Maximum allowed by Spotify API
			for {
				// Get tracks with pagination
				tracks, err := client.GetPlaylistTracksOpt(playlist.ID, &spotify.Options{
					Limit:  &limit,
					Offset: &offset,
				}, "")
				if err != nil {
					log.Printf("failed to get tracks for playlist %s: %v", playlist.Name, err)
					break
				}

				// Process tracks in this page
				for _, item := range tracks.Tracks {
					track := item.Track
					topTracksMap[string(track.ID)] = TrackInfo{
						ID:   string(track.ID),
						Name: track.Name,
					}
				}

				// Check if we've reached the end of the tracks
				if len(tracks.Tracks) < limit {
					break
				}

				offset += limit
			}
		}
	}

	fmt.Printf("Found %d unique tracks in top tracks playlists\n", len(topTracksMap))

	// Create a directory for CSV files
	if err := os.MkdirAll("playlists", 0755); err != nil {
		log.Fatalf("failed to create directory: %v", err)
	}

	// Create separate CSV files for user's playlists and other playlists
	userFilename := "playlists/user_playlists.csv"
	otherFilename := "playlists/other_playlists.csv"

	// Create and set up the user playlists file
	userFile, err := os.Create(userFilename)
	if err != nil {
		log.Fatalf("failed to create user playlists file: %v", err)
	}
	defer userFile.Close()

	// Create and set up the other playlists file if needed
	var otherFile *os.File
	if includeOtherPlaylists == "true" {
		otherFile, err = os.Create(otherFilename)
		if err != nil {
			log.Fatalf("failed to create other playlists file: %v", err)
		}
		defer otherFile.Close()
	}

	// Add UTF-8 BOM to both files
	userFile.Write([]byte{0xEF, 0xBB, 0xBF})
	if otherFile != nil {
		otherFile.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	userWriter := csv.NewWriter(userFile)
	defer userWriter.Flush()

	var otherWriter *csv.Writer
	if otherFile != nil {
		otherWriter = csv.NewWriter(otherFile)
		defer otherWriter.Flush()
	}

	// Write CSV headers
	headers := []string{"Playlist", "Track Name", "Artist(s)", "Album", "Release Date", "Release Year", "NotInTopTrackPlaylist"}
	if err := userWriter.Write(headers); err != nil {
		log.Fatalf("failed to write headers to user playlists file: %v", err)
	}
	if otherWriter != nil {
		if err := otherWriter.Write(headers); err != nil {
			log.Fatalf("failed to write headers to other playlists file: %v", err)
		}
	}

	// Process each playlist
	for _, playlist := range allPlaylists {
		// Determine which file to write to based on playlist ownership
		var writer *csv.Writer
		if playlist.Owner.ID != user.ID {
			if includeOtherPlaylists != "true" {
				fmt.Printf("Skipping playlist '%s' (Created by: %s)\n", playlist.Name, playlist.Owner.DisplayName)
				continue
			} else {
				fmt.Printf("Including playlist '%s' (Created by: %s)\n", playlist.Name, playlist.Owner.DisplayName)
				writer = otherWriter
			}
		} else {
			writer = userWriter
		}

		fmt.Printf("Processing playlist: %s (Created by: %s)\n", playlist.Name, playlist.Owner.DisplayName)

		// Process all tracks in the playlist with pagination
		offset := 0
		limit := 100 // Maximum allowed by Spotify API
		for {
			// Get tracks with pagination
			tracks, err := client.GetPlaylistTracksOpt(playlist.ID, &spotify.Options{
				Limit:  &limit,
				Offset: &offset,
			}, "")
			if err != nil {
				log.Printf("failed to get tracks for playlist %s: %v", playlist.Name, err)
				break
			}

			// Write tracks to CSV
			for _, item := range tracks.Tracks {
				track := item.Track
				artists := ""
				for i, artist := range track.Artists {
					if i > 0 {
						artists += ", "
					}
					artists += artist.Name
				}

				// Extract year from release date
				releaseYear := ""
				if track.Album.ReleaseDate != "" {
					// Handle different date formats (YYYY-MM-DD or YYYY)
					parts := strings.Split(track.Album.ReleaseDate, "-")
					if len(parts) > 0 {
						releaseYear = parts[0]
					}
				}

				// Check if track is from 2020-2025 and not in top tracks
				notInTopTracks := ""
				if releaseYear != "" {
					year := releaseYear
					if year >= startYear && year <= endYear {
						trackID := string(track.ID)
						if _, exists := topTracksMap[trackID]; !exists {
							notInTopTracks = "TRUE"
						}
					}
				}

				row := []string{
					playlist.Name,
					track.Name,
					artists,
					track.Album.Name,
					track.Album.ReleaseDate,
					releaseYear,
					notInTopTracks,
				}

				if err := writer.Write(row); err != nil {
					log.Printf("failed to write track %s for playlist %s: %v", track.Name, playlist.Name, err)
				}
			}

			// Check if we've reached the end of the tracks
			if len(tracks.Tracks) < limit {
				break
			}

			offset += limit
		}

		fmt.Printf("Completed playlist: %s\n", playlist.Name)
	}

	fmt.Println("All playlists have been processed!")
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

// normalizeQuotes replaces smart quotes with regular quotes
func normalizeQuotes(s string) string {
	s = strings.ReplaceAll(s, "\u2019", "'")  // Replace right single quotation mark with regular single quote
	s = strings.ReplaceAll(s, "\u2018", "'")  // Replace left single quotation mark with regular single quote
	s = strings.ReplaceAll(s, "\u201C", "\"") // Replace left double quotation mark with regular double quote
	s = strings.ReplaceAll(s, "\u201D", "\"") // Replace right double quotation mark with regular double quote
	return s
}
