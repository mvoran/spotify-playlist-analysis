# Spotify Playlist Analysis Tool

A Go application that analyzes your Spotify playlists to identify tracks from a specified year range that don't appear in your "Top Tracks" playlists.

## Motivation

Each year I build a "My Top Tracks of {Year}" Spotify playlist. However, Spotify doesn't have great UX for finding tracks by year, so I'm never 100% confident that I've found all eligible tracks in my playlists. What this tool does is analyze the Spotify playlists that I've created, compares the playlist contents to my "Top Tracks" playlists, and for any track that is not in a Top Tracks playlist (but released in a year covered by a Top Tracks playlsit) to flag it.

## Features

- Fetches all playlists from your Spotify account
- Identifies tracks from your specified year range (default: 2020-2025)
- Marks tracks that don't appear in your "Top Tracks" playlists
- Supports analyzing playlists created by other users (optional)
- Generates separate CSV files for your playlists and others' playlists
- Handles pagination for large playlists
- Normalizes smart quotes in playlist names
- Comprehensive logging with file rotation

## Prerequisites

- Go 1.16 or later
- A Spotify Developer account
- A Spotify application with appropriate credentials

## Setup

1. Create a Spotify Developer account at https://developer.spotify.com/dashboard
2. Create a new application in the Spotify Developer Dashboard
3. Add `http://localhost:8081/callback` to your application's Redirect URIs
4. Note down your application's Client ID and Client Secret

## Configuration

Create a `.env` file in the project root with the following variables:

```env
# Spotify API Credentials
SPOTIFY_CLIENT_ID=your_client_id_here
SPOTIFY_CLIENT_SECRET=your_client_secret_here

# Application Configuration
SPOTIFY_REDIRECT_URI=http://localhost:8081/callback
SPOTIFY_PORT=8081
SPOTIFY_TOP_TRACKS_PATTERN=your_pattern_here
SPOTIFY_START_YEAR=2020
SPOTIFY_END_YEAR=2025
SPOTIFY_INCLUDE_OTHER_PLAYLISTS=false
SPOTIFY_OVERWRITE_FILES=true

# Logging Configuration
SPOTIFY_LOG_FILE=logs/spotify-analysis.log
SPOTIFY_LOG_ROTATE_SIZE=10MB
SPOTIFY_LOG_KEEP_FILES=7
```

Replace:
- `your_client_id_here` with your Spotify application's Client ID
- `your_client_secret_here` with your Spotify application's Client Secret
- `your_pattern_here` with the pattern to identify your top tracks playlists (e.g., "jpizzle's top tracks of")
- `2020` with the first year of your top tracks range
- `2025` with the last year of your top tracks range
- `false` with `true` if you want to analyze playlists not created by you
- `true` with `false` if you don't want to overwrite existing CSV files
- `logs/spotify-analysis.log` with your preferred log file path
- `10MB` with your preferred log file size limit
- `7` with the number of old log files to keep

### Logging Configuration

The application provides comprehensive logging with the following features:
- Logs are written to both the terminal and a file
- Log files are automatically rotated when they reach the size limit
- Old log files are kept for historical reference
- Each log entry includes timestamp, file name, and line number
- Log files are stored in the `logs` directory by default

Default logging values:
- Log file: `logs/spotify-analysis.log`
- Rotate size: `10MB`
- Keep files: `7`

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/spotify-analysis.git
   cd spotify-analysis
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Usage

1. Run the program:
   ```bash
   go run main.go
   ```

2. Open your browser and log in to Spotify when prompted
3. The program will analyze your playlists and generate CSV files

## Output

The program generates CSV files in the `playlists` directory:
- `user_playlists.csv`: Contains tracks from playlists created by the authenticated user
- `other_playlists.csv`: Contains tracks from playlists created by other users (only generated if `SPOTIFY_INCLUDE_OTHER_PLAYLISTS=true`)

Each CSV file includes:
- UTF-8 BOM for proper Excel encoding
- All tracks from the respective playlists
- Special marking for tracks from the specified year range that don't appear in your top tracks playlists

Log files are stored in the `logs` directory:
- Current log file: `spotify-analysis.log`
- Rotated log files: `spotify-analysis-YYYY-MM-DD-HH-MM-SS.log`

## Notes

- The program handles pagination for both playlists and tracks
- Smart quotes in playlist names are automatically normalized
- By default, the program only processes playlists created by the authenticated user
- Set `SPOTIFY_INCLUDE_OTHER_PLAYLISTS=true` to analyze playlists created by other users
- Tracks from other users' playlists are saved to a separate CSV file
- Log files are automatically rotated when they reach the size limit
- Old log files are kept for historical reference

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 