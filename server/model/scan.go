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

func ScanString(row interface{ Scan(...any) error }) (string, error) {
	var value string
	err := row.Scan(&value)
	return value, err
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

func ScanDisplayConcert(row interface{ Scan(...any) error }) (DisplayConcert, error) {
	var concert DisplayConcert
	var date string
	var photoURL string
	var saleStart string
	err := row.Scan(
		&concert.ID,
		&concert.Name,
		&date,
		&concert.VenueID,
		&concert.ArtistID,
		&concert.URL,
		&photoURL,
		&concert.SeatmapURL,
		&saleStart,
		&concert.VenueName,
		&concert.ArtistName,
	)
	if err != nil {
		return DisplayConcert{}, err
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
	err := row.Scan(
		&passkey.CredentialID,
		&passkey.PublicKey,
		&passkey.SignCount,
		&passkey.UserPresent,
		&passkey.UserVerified,
		&passkey.BackupEligible,
		&passkey.BackupState,
	)
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

func ScanProfileWT(row interface{ Scan(...any) error }) (ProfileWT, error) {
	var wt ProfileWT
	var date string
	var photoURL string
	var saleStart string
	err := row.Scan(
		&wt.Type,
		&wt.Concert.ID,
		&wt.Concert.Name,
		&date,
		&wt.Concert.VenueID,
		&wt.Concert.ArtistID,
		&wt.Concert.URL,
		&photoURL,
		&wt.Concert.SeatmapURL,
		&saleStart,
		&wt.Concert.VenueName,
		&wt.Concert.ArtistName,
	)
	if err != nil {
		return ProfileWT{}, err
	}
	wt.Concert.Date = parseTime(date)
	if photoURL != "" {
		wt.Concert.Photos = []string{photoURL}
	}
	wt.Concert.SaleStartDateTime = parseTime(saleStart)
	return wt, nil
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, value)
	return t
}
