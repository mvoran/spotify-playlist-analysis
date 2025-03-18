# Spotify Playlist Analysis

This Go program analyzes your Spotify playlists and identifies tracks from a specified year range that don't appear in your "Top Tracks" playlists.

## Motivation

Each year I build a "My Top Tracks of {Year}" Spotify playlist. However, Spotify doesn't have great UX for finding tracks by year, so I'm never 100% confident that I've found all eligible tracks in my playlists. What this tool does is analyze the Spotify playlists that I've created, compares the playlist contents to my "Top Tracks" playlists, and for any track that is not in a Top Tracks playlist (but released in a year covered by a Top Tracks playlsit) to flag it.

## Features

- Fetches all your Spotify playlists
- Identifies your "Top Tracks" playlists based on a configurable pattern
- Processes all tracks in each playlist
- Exports track information to a CSV file with the following columns:
  - Playlist name
  - Track name
  - Artist(s)
  - Album
  - Release date
  - Release year
  - NotInTopTrackPlaylist (TRUE if the track is from the specified year range and not in any top tracks playlist)

## Prerequisites

- Go 1.16 or later
- A Spotify account
- A Spotify Developer account and application

## Setup

1. Create a Spotify Developer account at https://developer.spotify.com/
2. Create a new application in the Spotify Developer Dashboard
3. Add `http://localhost:8081/callback` as a Redirect URI in your application settings
4. Copy your Client ID and Client Secret

## Configuration

Create a `.env` file in the project root with the following variables:

```env
SPOTIFY_CLIENT_ID=your_client_id_here
SPOTIFY_CLIENT_SECRET=your_client_secret_here
SPOTIFY_REDIRECT_URI=http://localhost:8081/callback
SPOTIFY_PORT=8081
SPOTIFY_TOP_TRACKS_PATTERN=your_pattern_here
SPOTIFY_START_YEAR=2020
SPOTIFY_END_YEAR=2025
SPOTIFY_INCLUDE_OTHER_PLAYLISTS=false
```

Replace:
- `your_client_id_here` with your Spotify application's Client ID
- `your_client_secret_here` with your Spotify application's Client Secret
- `your_pattern_here` with the pattern to identify your top tracks playlists (e.g., "jpizzle's top tracks of")
- `2020` with the first year of your top tracks range
- `2025` with the last year of your top tracks range
- `false` with `true` if you want to analyze playlists not created by you

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Create and configure the `.env` file as described above

## Usage

1. Run the program:
   ```bash
   go run main.go
   ```
2. A browser window will open automatically for Spotify authentication
3. After authentication, the program will:
   - Fetch all your playlists
   - Process tracks from each playlist
   - Generate a CSV file in the `playlists` directory

## Output

The program generates CSV files in the `playlists` directory:
- `user_playlists.csv`: Contains tracks from playlists created by the authenticated user
- `other_playlists.csv`: Contains tracks from playlists created by other users (only generated if `SPOTIFY_INCLUDE_OTHER_PLAYLISTS=true`)

Each CSV file includes:
- UTF-8 BOM for proper Excel encoding
- All tracks from the respective playlists
- Special marking for tracks from the specified year range that don't appear in your top tracks playlists

## Notes

- The program handles pagination for both playlists and tracks
- Smart quotes in playlist names are automatically normalized
- By default, the program only processes playlists created by the authenticated user
- Set `SPOTIFY_INCLUDE_OTHER_PLAYLISTS=true` to analyze playlists created by other users
- Tracks from other users' playlists are saved to a separate CSV file 