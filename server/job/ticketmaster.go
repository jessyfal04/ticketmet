package job

import (
	"context"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"net/url"
	"server/model"
	"strconv"
	"time"
)

const (
	ticketmasterURL         = "https://app.ticketmaster.com/discovery/v2/events.json"
	ticketmasterHTTPTimeout = 10 * time.Second
	ticketmasterPageSize    = 200
	ticketmasterMaxPages    = 5
	ticketmasterMusicClass  = "music"
)

var ticketmasterCountries = []string{"FR", "DE", "BE", "NL", "LU", "CH", "AD", "GB", "IT", "MC", "ES"}

// TicketmasterSyncStats summarizes one sync pass.
type TicketmasterSyncStats struct {
	StartedAt time.Time
	Fetched   int
	Saved     int
	Skipped   int
}

// Stats to string for log
func (s TicketmasterSyncStats) String() string {
	return fmt.Sprintf(
		"[Ticketmaster] fetched=%d saved=%d skipped=%d duration=%s",
		s.Fetched, s.Saved, s.Skipped, time.Since(s.StartedAt).Round(time.Millisecond),
	)
}

// Start the periodic Ticketmaster sync
func RunTicketmaster(ctx context.Context, dbChan chan<- DBRequest, apiKey string, interval time.Duration) {
	runEvery(ctx, interval, func() {
		stats, err := SyncTicketmaster(ctx, dbChan, apiKey)
		if err != nil {
			log.Printf("[Ticketmaster] failed: %v (%s)", err, stats)
			return
		}
		log.Print(stats)
	})
}

// Response of Ticketmaster events API
type ticketmasterResp_Page struct {
	Embedded struct {
		Events []ticketmasterResp_Event `json:"events"`
	} `json:"_embedded"`
	Page struct {
		Number     int `json:"number"`
		TotalPages int `json:"totalPages"`
	} `json:"page"`
}

// One Ticketmaster event
type ticketmasterResp_Event struct {
	ID                            string `json:"id"`
	Name                          string `json:"name"`
	URL                           string `json:"url"`
	PublicVisibilityStartDateTime string `json:"publicVisibilityStartDateTime"`
	Images                        []struct {
		Ratio  string `json:"ratio"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"images"`
	Dates struct {
		Start struct {
			LocalDate string `json:"localDate"`
			LocalTime string `json:"localTime"`
			DateTime  string `json:"dateTime"`
		} `json:"start"`
	} `json:"dates"`
	Sales struct {
		Public struct {
			StartDateTime string `json:"startDateTime"`
		} `json:"public"`
	} `json:"sales"`
	Seatmap struct {
		StaticURL string `json:"staticUrl"`
	} `json:"seatmap"`
	Embedded struct {
		Venues []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			City struct {
				Name string `json:"name"`
			} `json:"city"`
			Country struct {
				CountryCode string `json:"countryCode"`
			} `json:"country"`
		} `json:"venues"`
		Attractions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"attractions"`
	} `json:"_embedded"`
}

// One sync for Ticketmaster
func SyncTicketmaster(ctx context.Context, dbChan chan<- DBRequest, apiKey string) (TicketmasterSyncStats, error) {
	// Stats with start time
	stats := TicketmasterSyncStats{StartedAt: time.Now().UTC()}

	// Get the current visibility watermark to fetch only new events
	maxVisibility, err := SqlScanOne(ctx, dbChan, `
		SELECT max_visibility
		FROM sync_ticketmaster
		WHERE id = 1`,
		model.ScanTime)
	if err != nil {
		return stats, err
	}

	// Fetch the events from Ticketmaster API
	events, err := fetchTicketmasterEvents(ctx, apiKey, maxVisibility)
	if err != nil {
		return stats, err
	}
	stats.Fetched = len(events)

	// Save the events in the db
	// newMaxVisibility will track the most recent public visibility
	newMaxVisibility := maxVisibility
	for _, event := range events {
		saved, err := saveEvent(ctx, dbChan, event)
		if err != nil {
			return stats, err
		}

		// Update stats and newMaxVisibility if possible
		if saved {
			stats.Saved++
			visibility := model.ParseTimeText(event.PublicVisibilityStartDateTime)
			if visibility.After(newMaxVisibility) {
				newMaxVisibility = visibility
			}
		} else {
			stats.Skipped++
		}
	}

	// Update the max visibility
	if err := SqlExec(ctx, dbChan, `
		INSERT INTO sync_ticketmaster (id, max_visibility)
		VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET max_visibility = excluded.max_visibility`,
		newMaxVisibility.UTC().Format(time.RFC3339)); err != nil {
		return stats, err
	}

	return stats, nil
}

// Fetch all Ticketmaster pages for the configured countries.
func fetchTicketmasterEvents(ctx context.Context, apiKey string, maxVisibility time.Time) ([]ticketmasterResp_Event, error) {
	// Event list
	var events []ticketmasterResp_Event

	// Client with timeout.
	client := &http.Client{Timeout: ticketmasterHTTPTimeout}

	// For each country
	for _, country := range ticketmasterCountries {
		log.Printf("[Ticketmaster] fetching country=%s", country)

		// For each page until the max
		for page := 0; page < ticketmasterMaxPages; page++ {
			// Build the query.
			q := url.Values{}
			q.Set("apikey", apiKey)
			q.Set("classificationName", ticketmasterMusicClass)
			q.Set("countryCode", country)
			q.Set("size", strconv.Itoa(ticketmasterPageSize))
			q.Set("page", strconv.Itoa(page))
			if !maxVisibility.IsZero() { // if we have a visibility, we filter
				q.Set("publicVisibilityStartDateTime", maxVisibility.UTC().Format(time.RFC3339))
			}

			// Response payload.
			var payload ticketmasterResp_Page
			// Make the request.
			if err := getJSONQuery(ctx, client, ticketmasterURL, q, nil, &payload, false); err != nil {
				return nil, err
			}

			// Adding to the agg list
			events = append(events, payload.Embedded.Events...)

			// If we are at the last page, we can stop fetching for this country.
			if payload.Page.Number+1 >= payload.Page.TotalPages {
				break
			}
		}
	}

	return events, nil
}

// Save one Ticketmaster event into the database
func saveEvent(ctx context.Context, dbChan chan<- DBRequest, event ticketmasterResp_Event) (bool, error) {
	// We reject events without usable event/venue/attraction data
	if len(event.Embedded.Venues) == 0 || len(event.Embedded.Attractions) == 0 {
		return false, nil
	}
	venue := event.Embedded.Venues[0]
	artist := event.Embedded.Attractions[0]
	venueOK, venueID := stableID(venue.ID)
	artistOK, artistID := stableID(artist.ID)
	concertOK, concertID := stableID(event.ID)
	if event.Name == "" || venue.Name == "" || artist.Name == "" || !venueOK || !artistOK || !concertOK {
		return false, nil
	}

	// Try saving the venue, can update if it already exists, if already here error
	if err := SqlExec(ctx, dbChan, `
		INSERT INTO venues (id, name, city, country)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			city = excluded.city,
			country = excluded.country`,
		venueID, venue.Name, venue.City.Name, venue.Country.CountryCode); err != nil {
		return false, err
	}

	// Try saving the artist, can update if it already exists, if already here error
	if err := SqlExec(ctx, dbChan, `
		INSERT INTO artists (id, name)
		VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name`,
		artistID, artist.Name); err != nil {
		return false, err
	}

	dateValue := ticketmasterDateValue(event)

	saleValue := ticketmasterSaleValue(event.Sales.Public.StartDateTime)

	// Save the concert details
	if err := SqlExec(ctx, dbChan, `
		INSERT INTO concerts (id, name, date, venue_id, artist_id, url, photo_url, seatmap_url, sale_start_datetime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			date = excluded.date,
			venue_id = excluded.venue_id,
			artist_id = excluded.artist_id,
			url = excluded.url,
		photo_url = excluded.photo_url,
		seatmap_url = excluded.seatmap_url,
		sale_start_datetime = excluded.sale_start_datetime`,
		concertID, event.Name, dateValue, venueID, artistID, event.URL, ticketmasterBestImage(event), event.Seatmap.StaticURL, saleValue); err != nil {
		return false, err
	}
	return true, nil
}

// Helpeur Ticketmaster

// Ticketmaster can return placeholder sale dates like 1900-01-01.
func ticketmasterSaleValue(value string) string {
	parsed := model.ParseTimeText(value)
	if parsed.IsZero() || parsed.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return ""
	}
	return parsed.UTC().Format(time.RFC3339)
}

// Ticketmaster can return the event date in a few different positions
func ticketmasterDateValue(event ticketmasterResp_Event) string {
	// We keep the first valid value and normalize everything to RFC3339 UTC.

	// dates.start.dateTime
	if parsed := model.ParseTimeText(event.Dates.Start.DateTime); !parsed.IsZero() {
		return parsed.UTC().Format(time.RFC3339)
	}

	// dates.start.localDate + dates.start.localTime
	if event.Dates.Start.LocalDate != "" && event.Dates.Start.LocalTime != "" {
		value := event.Dates.Start.LocalDate + "T" + event.Dates.Start.LocalTime
		for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04"} {
			if parsed, err := time.Parse(layout, value); err == nil {
				return parsed.UTC().Format(time.RFC3339)
			}
		}
	}

	// dates.start.dateTime
	if parsed, err := time.Parse("2006-01-02", event.Dates.Start.LocalDate); err == nil {
		return parsed.UTC().Format(time.RFC3339)
	}

	return ""
}

// Pick the best image Ticketmaster gives us
func ticketmasterBestImage(event ticketmasterResp_Event) string {
	// We pick the image with the higher score among every available one
	best := ""
	bestScore := -1
	for _, image := range event.Images {
		// Need an URL
		if image.URL == "" {
			continue
		}

		// Score is by image size, + *10 bonus if it's 16:9 ratio
		score := image.Width * image.Height
		if image.Ratio == "16_9" {
			score *= 10
		}

		// If it's the current best, we keep it
		if score > bestScore {
			bestScore = score
			best = image.URL
		}
	}
	return best
}

// SQLite ID from Ticketmaster ID
func stableID(value string) (bool, int) {
	if value == "" {
		return false, 42
	}
	// We use a hash to convert the Ticketmaster string id to a SQLite int id.
	id := int(crc32.ChecksumIEEE([]byte(value)))
	return true, id
}
