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

func ScanUser(row interface{ Scan(...any) error }) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	return user, err
}

func ScanPasskey(row interface{ Scan(...any) error }) (Passkey, error) {
	var passkey Passkey
	err := row.Scan(&passkey.CredentialID, &passkey.PublicKey, &passkey.SignCount)
	return passkey, err
}

func ScanPublicPasskey(row interface{ Scan(...any) error }) (PublicPasskey, error) {
	passkey, err := ScanPasskey(row)
	return passkey.Public(), err
}

func ScanWebAuthnChallenge(row interface{ Scan(...any) error }) (WebAuthnChallengeRow, error) {
	var challenge WebAuthnChallengeRow
	err := row.Scan(&challenge.ID, &challenge.SessionData)
	return challenge, err
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, value)
	return t
}
