package job

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/url"
	"time"

	"server/model"
)

const (
	ticketmasterAPIKey         = "REDACTED"
	ticketmasterCountryCode    = "DE"
	ticketmasterClassification = "music"
)

type ticketmasterResponse struct {
	Embedded ticketmasterEmbedded `json:"_embedded"`
}

type ticketmasterEmbedded struct {
	Events []ticketmasterEvent `json:"events"`
}

type ticketmasterEvent struct {
	ID                            string                 `json:"id"`
	Name                          string                 `json:"name"`
	URL                           string                 `json:"url"`
	Images                        []ticketmasterImage    `json:"images"`
	Dates                         ticketmasterDates      `json:"dates"`
	Sales                         ticketmasterSales      `json:"sales"`
	PublicVisibilityStartDateTime string                 `json:"publicVisibilityStartDateTime"`
	Embedded                      ticketmasterEventEmbed `json:"_embedded"`
}

type ticketmasterImage struct {
	URL string `json:"url"`
}

type ticketmasterDates struct {
	Start ticketmasterStart `json:"start"`
}

type ticketmasterStart struct {
	LocalDate string `json:"localDate"`
	LocalTime string `json:"localTime"`
	DateTime  string `json:"dateTime"`
}

type ticketmasterSales struct {
	Public ticketmasterSalesPublic `json:"public"`
}

type ticketmasterSalesPublic struct {
	StartDateTime string `json:"startDateTime"`
}

type ticketmasterEventEmbed struct {
	Venues      []ticketmasterVenue      `json:"venues"`
	Attractions []ticketmasterAttraction `json:"attractions"`
}

type ticketmasterVenue struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	City    ticketmasterCity    `json:"city"`
	Country ticketmasterCountry `json:"country"`
}

type ticketmasterCity struct {
	Name string `json:"name"`
}

type ticketmasterCountry struct {
	CountryCode string `json:"countryCode"`
}

type ticketmasterAttraction struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func SyncTicketmaster(data model.DataSet) (model.DataSet, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	reqURL := ticketmasterURL(data.Sync.LastPublicVisibilityStartDateTime)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return data, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("ticketmaster sync failed: status=%d", resp.StatusCode)
	}

	var payload ticketmasterResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return data, err
	}

	updated := data
	venueIndex := make(map[int]int, len(updated.Venues))
	for i, venue := range updated.Venues {
		venueIndex[venue.ID] = i
	}
	artistIndex := make(map[int]int, len(updated.Artists))
	for i, artist := range updated.Artists {
		artistIndex[artist.ID] = i
	}
	concertIndex := make(map[int]int, len(updated.Concerts))
	for i, concert := range updated.Concerts {
		concertIndex[concert.ID] = i
	}

	maxVisibility := data.Sync.LastPublicVisibilityStartDateTime

	for _, event := range payload.Embedded.Events {
		venueID := 0
		if len(event.Embedded.Venues) > 0 {
			venue := event.Embedded.Venues[0]
			venueID = stableID(venue.ID)
			if venueID != 0 {
				entry := model.Venue{
					ID:      venueID,
					Name:    venue.Name,
					City:    venue.City.Name,
					Country: venue.Country.CountryCode,
				}
				if idx, ok := venueIndex[venueID]; ok {
					updated.Venues[idx] = entry
				} else {
					venueIndex[venueID] = len(updated.Venues)
					updated.Venues = append(updated.Venues, entry)
				}
			}
		}

		artistID := 0
		if len(event.Embedded.Attractions) > 0 {
			artist := event.Embedded.Attractions[0]
			artistID = stableID(artist.ID)
			if artistID != 0 {
				entry := model.Artist{
					ID:   artistID,
					Name: artist.Name,
				}
				if idx, ok := artistIndex[artistID]; ok {
					updated.Artists[idx] = entry
				} else {
					artistIndex[artistID] = len(updated.Artists)
					updated.Artists = append(updated.Artists, entry)
				}
			}
		}

		eventID := stableID(event.ID)
		if eventID != 0 {
			saleStart := time.Time{}
			if parsed, ok := parseRFC3339(event.Sales.Public.StartDateTime); ok {
				saleStart = parsed
			}
			entry := model.Concert{
				ID:                eventID,
				Name:              event.Name,
				Date:              parseEventStart(event.Dates.Start),
				VenueID:           venueID,
				ArtistID:          artistID,
				URL:               event.URL,
				Photos:            extractImageURLs(event.Images),
				SaleStartDateTime: saleStart,
			}
			if idx, ok := concertIndex[eventID]; ok {
				updated.Concerts[idx] = entry
			} else {
				concertIndex[eventID] = len(updated.Concerts)
				updated.Concerts = append(updated.Concerts, entry)
			}
		}

		if visibility, ok := parseRFC3339(event.PublicVisibilityStartDateTime); ok && visibility.After(maxVisibility) {
			maxVisibility = visibility
		}
	}

	updated.Sync.LastPublicVisibilityStartDateTime = maxVisibility

	return updated, nil
}

func ticketmasterURL(lastVisibility time.Time) string {
	u := url.URL{
		Scheme: "https",
		Host:   "app.ticketmaster.com",
		Path:   "/discovery/v2/events.json",
	}
	q := url.Values{}
	q.Set("apikey", ticketmasterAPIKey)
	q.Set("countryCode", ticketmasterCountryCode)
	q.Set("classificationName", ticketmasterClassification)
	if !lastVisibility.IsZero() {
		q.Set("publicVisibilityStartDateTime", lastVisibility.UTC().Format(time.RFC3339))
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func extractImageURLs(images []ticketmasterImage) []string {
	if len(images) == 0 {
		return nil
	}
	urls := make([]string, 0, len(images))
	for _, image := range images {
		if image.URL != "" {
			urls = append(urls, image.URL)
		}
	}
	return urls
}

func parseEventStart(start ticketmasterStart) time.Time {
	if start.DateTime != "" {
		if parsed, ok := parseRFC3339(start.DateTime); ok {
			return parsed
		}
	}

	if start.LocalDate != "" && start.LocalTime != "" {
		dateTime := start.LocalDate + "T" + start.LocalTime
		for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04"} {
			if parsed, err := time.Parse(layout, dateTime); err == nil {
				return parsed
			}
		}
	}

	if start.LocalDate != "" {
		if parsed, err := time.Parse("2006-01-02", start.LocalDate); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

func parseRFC3339(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
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
