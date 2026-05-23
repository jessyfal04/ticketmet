package job

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"server/model"
	"strings"
	"time"
)

const (
	setlistFMURL         = "https://api.setlist.fm/rest/1.0/search/setlists"
	setlistFMHTTPTimeout = 12 * time.Second
)

// SetlistRequest & SetlistResult for communication
type SetlistRequest struct {
	ConcertID int
	Ret       chan SetlistResult
}

type SetlistResult struct {
	Songs []string
	Err   error
}

// Run the Setlist.fm server to handle setlist requests
func RunSetlistFMServer(ctx context.Context, dbChan chan<- DBRequest, apiKey string, c chan SetlistRequest) {
	go runChan(ctx, c, func(req SetlistRequest) {
		result, err := GetConcertSetlist(ctx, dbChan, apiKey, req.ConcertID)
		if err != nil {
			log.Printf("[setlistfm] failed concert_id=%d: %v", req.ConcertID, err)
			result.Err = err
		}
		select {
		case req.Ret <- result:
		case <-ctx.Done():
		}
	})
}

func GetConcertSetlist(ctx context.Context, dbChan chan<- DBRequest, apiKey string, concertID int) (SetlistResult, error) {
	// We try to get the cached setlist first
	cached, err := SqlScanList(ctx, dbChan, `
		SELECT song_name
		FROM setlists
		WHERE concert_id = ?
		ORDER BY song_order`,
		model.ScanString, concertID)
	if err != nil {
		return SetlistResult{}, err
	}
	if len(cached) > 0 {
		return SetlistResult{Songs: cached}, nil
	}

	// We check if there is an API key
	if apiKey == "" {
		log.Printf("[setlistfm] disabled: SETLISTFM_API_KEY is empty")
		return SetlistResult{Songs: cached}, nil
	}

	// We get the concert details
	artistName, err := SqlScanOne(ctx, dbChan, `
		SELECT a.name
		FROM concerts c
		JOIN artists a ON a.id = c.artist_id
		WHERE c.id = ?`,
		model.ScanString, concertID)
	if err != nil {
		return SetlistResult{}, err
	}

	// We fetch the songs from Setlist.fm
	songs, err := fetchSongs(ctx, apiKey, artistName)
	if err != nil {
		return SetlistResult{}, err
	}

	// No songs found, we return the empty setlist
	if len(songs) == 0 {
		return SetlistResult{}, nil
	}

	// We save the setlist in the database for future requests
	for i, song := range songs {
		if err := SqlExec(ctx, dbChan, `
			INSERT INTO setlists (concert_id, song_order, song_name)
			VALUES (?, ?, ?)`, concertID, i+1, song); err != nil {
			return SetlistResult{}, err
		}
	}

	// We return the setlist
	log.Printf("[setlistfm] saved setlist concert_id=%d songs=%d", concertID, len(songs))
	return SetlistResult{Songs: songs}, nil
}

// API Setlist.fm

// Response of Setlist.fm search API - list of Setlists
type setlistFMResp_Setlists struct {
	Setlists []setlistFMResp_Setlist `json:"setlist"`
}

// Response of Setlist.fm search API - one Setlist with songs
type setlistFMResp_Setlist struct {
	Sets struct {
		Set []struct {
			Song []struct {
				Name string `json:"name"`
			} `json:"song"`
		} `json:"set"`
	} `json:"sets"`
}

// Fetch the songs from Setlist.fm using only the artist name.
func fetchSongs(ctx context.Context, apiKey string, artistName string) ([]string, error) {
	// Build the query
	q := url.Values{}
	q.Set("p", "1")
	q.Set("artistName", strings.TrimSpace(artistName))

	// Client with timeout
	client := &http.Client{Timeout: setlistFMHTTPTimeout}

	// Headers with API key
	headers := map[string]string{
		"Accept":     "application/json",
		"x-api-key":  apiKey,
		"User-Agent": "ticketmet/1.0",
	}

	// Response payload
	var payload setlistFMResp_Setlists

	// Make the request
	if err := getJSONQuery(ctx, client, setlistFMURL, q, headers, &payload, true); err != nil {
		return nil, fmt.Errorf("[setlistfm]: %w", err)
	}

	// Extract the songs from the response
	for _, setlist := range payload.Setlists {
		songs := setlistSongs(setlist)

		// If we found songs, we return them, otherwise we try the next setlist.
		if len(songs) > 0 {
			return songs, nil
		}
	}

	return nil, nil
}

// Take a setlist.fm setlist response and extract the songs in order.
func setlistSongs(setlist setlistFMResp_Setlist) []string {
	// Get the songs in order and add them to the list
	var songs []string
	for _, set := range setlist.Sets.Set {
		for _, song := range set.Song {
			songs = append(songs, strings.TrimSpace(song.Name))
		}
	}
	return songs
}
