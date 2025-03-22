package spotify

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/mikev/spotify-analysis/pkg/config"
	"github.com/zmb3/spotify"
)

var (
	auth  spotify.Authenticator
	ch    = make(chan *spotify.Client)
	state = "some-random-state-key"
)

// Client wraps the Spotify client with additional functionality
type Client struct {
	*spotify.Client
	server     *http.Server
	serverWg   sync.WaitGroup
	serverOnce sync.Once
}

// NewClient creates a new authenticated Spotify client
func NewClient(cfg *config.Config) (*Client, error) {
	// Check and cleanup port before starting
	if err := checkAndCleanupPort(cfg.Port); err != nil {
		return nil, fmt.Errorf("failed to cleanup port: %v", err)
	}

	// Initialize the authenticator
	auth = spotify.NewAuthenticator(cfg.RedirectURI, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadCollaborative)
	auth.SetAuthInfo(cfg.ClientID, cfg.ClientSecret)

	// Create a new server with timeout
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Create the client
	client := &Client{
		Client: nil,
		server: server,
	}

	// Start local server to receive the callback
	http.HandleFunc("/callback", client.completeAuth)

	// Start server in a goroutine
	client.serverWg.Add(1)
	go func() {
		defer client.serverWg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Open the browser for authentication
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	openBrowser(url)

	// Wait for auth to complete with timeout
	select {
	case spotifyClient := <-ch:
		client.Client = spotifyClient
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timed out after 5 minutes")
	}

	return client, nil
}

// Cleanup properly shuts down the HTTP server
func (c *Client) Cleanup() {
	c.serverOnce.Do(func() {
		if c.server != nil {
			if err := c.server.Close(); err != nil {
				log.Printf("Error closing server: %v", err)
			}
			c.serverWg.Wait()
		}
	})
}

func (c *Client) completeAuth(w http.ResponseWriter, r *http.Request) {
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
