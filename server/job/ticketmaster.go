package job

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	ticketmasterURL        = "https://app.ticketmaster.com/discovery/v2/events.json"
	ticketmasterSyncEvery  = 5 * time.Minute
	ticketmasterPageSize   = 200
	ticketmasterMaxPages   = 5
	ticketmasterMusicClass = "music"
)

var ticketmasterCountries = []string{"DE", "FR"}

type ticketmasterResponse struct {
	Embedded struct {
		Events []ticketmasterEvent `json:"events"`
	} `json:"_embedded"`
	Page struct {
		Number     int `json:"number"`
		TotalPages int `json:"totalPages"`
	} `json:"page"`
}

type ticketmasterEvent struct {
	ID                            string `json:"id"`
	Name                          string `json:"name"`
	URL                           string `json:"url"`
	PublicVisibilityStartDateTime string `json:"publicVisibilityStartDateTime"`
	Images                        []struct {
		URL string `json:"url"`
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

func StartTicketmaster(db *sql.DB, apiKey string) {
	if apiKey == "" {
		log.Printf("Ticketmaster sync disabled: TICKETMASTER_API_KEY is empty")
		return
	}

	go func() {
		for {
			if err := SyncTicketmaster(context.Background(), db, apiKey); err != nil {
				log.Printf("Ticketmaster sync failed: %v", err)
			} else {
				log.Printf("Ticketmaster sync done")
			}
			time.Sleep(ticketmasterSyncEvery)
		}
	}()
}

func SyncTicketmaster(ctx context.Context, db *sql.DB, apiKey string) error {
	lastSync, err := readLastSync(ctx, db)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	maxVisibility := lastSync
	for _, country := range ticketmasterCountries {
		countryMax, err := syncCountry(ctx, tx, apiKey, country, lastSync)
		if err != nil {
			return err
		}
		if countryMax.After(maxVisibility) {
			maxVisibility = countryMax
		}
	}

	// Ticketmaster does not always send publicVisibilityStartDateTime.
	// If nothing better was found, remember "now" to avoid full resync forever.
	if !maxVisibility.After(lastSync) {
		maxVisibility = time.Now().UTC()
	}
	if err := saveLastSync(ctx, tx, maxVisibility); err != nil {
		return err
	}

	return tx.Commit()
}

func syncCountry(ctx context.Context, tx *sql.Tx, apiKey string, country string, lastSync time.Time) (time.Time, error) {
	maxVisibility := lastSync
	client := &http.Client{Timeout: 10 * time.Second}

	for page := 0; page < ticketmasterMaxPages; page++ {
		payload, err := fetchPage(ctx, client, apiKey, country, lastSync, page)
		if err != nil {
			return maxVisibility, err
		}

		for _, event := range payload.Embedded.Events {
			visibility, saved, err := saveEvent(ctx, tx, event)
			if err != nil {
				return maxVisibility, err
			}
			if saved && visibility.After(maxVisibility) {
				maxVisibility = visibility
			}
		}

		if payload.Page.Number+1 >= payload.Page.TotalPages {
			break
		}
	}

	return maxVisibility, nil
}

func fetchPage(ctx context.Context, client *http.Client, apiKey string, country string, lastSync time.Time, page int) (ticketmasterResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildURL(apiKey, country, lastSync, page), nil)
	if err != nil {
		return ticketmasterResponse{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return ticketmasterResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ticketmasterResponse{}, fmt.Errorf("%s page %d: ticketmaster status=%d", country, page, resp.StatusCode)
	}

	var payload ticketmasterResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ticketmasterResponse{}, err
	}
	return payload, nil
}

func saveEvent(ctx context.Context, tx *sql.Tx, event ticketmasterEvent) (time.Time, bool, error) {
	// Empty venues/artists make filters unusable, so do not keep those events.
	if len(event.Embedded.Venues) == 0 || len(event.Embedded.Attractions) == 0 {
		return time.Time{}, false, nil
	}
	venue := event.Embedded.Venues[0]
	artist := event.Embedded.Attractions[0]
	if venue.ID == "" || venue.Name == "" || artist.ID == "" || artist.Name == "" {
		return time.Time{}, false, nil
	}

	venueID := stableID(venue.ID)
	artistID := stableID(artist.ID)
	concertID := stableID(event.ID)
	if venueID == 0 || artistID == 0 || concertID == 0 {
		return time.Time{}, false, nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO venues (id, name, city, country)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			city = excluded.city,
			country = excluded.country`,
		venueID, venue.Name, venue.City.Name, venue.Country.CountryCode); err != nil {
		return time.Time{}, false, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO artists (id, name)
		VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name`,
		artistID, artist.Name); err != nil {
		return time.Time{}, false, err
	}

	if _, err := tx.ExecContext(ctx, `
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
		concertID,
		event.Name,
		formatTime(eventDate(event)),
		venueID,
		artistID,
		event.URL,
		firstImage(event),
		event.Seatmap.StaticURL,
		formatTime(saleDate(event.Sales.Public.StartDateTime))); err != nil {
		return time.Time{}, false, err
	}

	visibility, _ := parseTime(event.PublicVisibilityStartDateTime)
	return visibility, true, nil
}

func readLastSync(ctx context.Context, db *sql.DB) (time.Time, error) {
	var value string
	err := db.QueryRowContext(ctx, `
		SELECT last_public_visibility_start_datetime
		FROM sync_ticketmaster
		WHERE id = 1`).Scan(&value)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	parsed, _ := parseTime(value)
	return parsed, nil
}

func saveLastSync(ctx context.Context, tx *sql.Tx, value time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sync_ticketmaster (id, last_public_visibility_start_datetime)
		VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET last_public_visibility_start_datetime = excluded.last_public_visibility_start_datetime`,
		value.UTC().Format(time.RFC3339))
	return err
}

func buildURL(apiKey string, country string, lastSync time.Time, page int) string {
	u, _ := url.Parse(ticketmasterURL)
	q := u.Query()
	q.Set("apikey", apiKey)
	q.Set("classificationName", ticketmasterMusicClass)
	q.Set("countryCode", country)
	q.Set("size", strconv.Itoa(ticketmasterPageSize))
	q.Set("page", strconv.Itoa(page))
	if !lastSync.IsZero() {
		q.Set("publicVisibilityStartDateTime", lastSync.UTC().Format(time.RFC3339))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func eventDate(event ticketmasterEvent) time.Time {
	if parsed, ok := parseTime(event.Dates.Start.DateTime); ok {
		return parsed
	}

	if event.Dates.Start.LocalDate != "" && event.Dates.Start.LocalTime != "" {
		value := event.Dates.Start.LocalDate + "T" + event.Dates.Start.LocalTime
		for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04"} {
			if parsed, err := time.Parse(layout, value); err == nil {
				return parsed
			}
		}
	}

	if event.Dates.Start.LocalDate != "" {
		if parsed, err := time.Parse("2006-01-02", event.Dates.Start.LocalDate); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

func saleDate(value string) time.Time {
	parsed, _ := parseTime(value)
	// Ticketmaster sometimes returns placeholder dates like 1900-01-01.
	if parsed.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return time.Time{}
	}
	return parsed
}

func parseTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil
}

func firstImage(event ticketmasterEvent) string {
	// Ticketmaster sends the same artwork many times in different sizes.
	for _, image := range event.Images {
		if image.URL != "" {
			return image.URL
		}
	}
	return ""
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func stableID(value string) int {
	if value == "" {
		return 0
	}
	id := int(crc32.ChecksumIEEE([]byte(value)))
	if id == 0 {
		return 1
	}
	return id
}
