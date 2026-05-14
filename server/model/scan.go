package model

import "time"

func ScanArtist(row interface{ Scan(...any) error }) (Artist, error) {
	var artist Artist
	err := row.Scan(&artist.ID, &artist.Name)
	return artist, err
}

func ScanVenue(row interface{ Scan(...any) error }) (Venue, error) {
	var venue Venue
	err := row.Scan(&venue.ID, &venue.Name, &venue.City, &venue.Country)
	return venue, err
}

func ScanConcert(row interface{ Scan(...any) error }) (Concert, error) {
	var concert Concert
	var date string
	var photoURL string
	var saleStart string
	err := row.Scan(&concert.ID, &concert.Name, &date, &concert.VenueID, &concert.ArtistID, &concert.URL, &photoURL, &concert.SeatmapURL, &saleStart)
	if err != nil {
		return Concert{}, err
	}
	concert.Date = parseTime(date)
	if photoURL != "" {
		concert.Photos = []string{photoURL}
	}
	concert.SaleStartDateTime = parseTime(saleStart)
	return concert, nil
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, value)
	return t
}
