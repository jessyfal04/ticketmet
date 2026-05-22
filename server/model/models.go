package model

import (
	"time"
)

type Concert struct {
	ID                int
	Name              string
	Date              time.Time
	VenueID           int
	ArtistID          int
	URL               string
	SeatmapURL        string
	Photos            []string
	SaleStartDateTime time.Time
}

type DisplayConcert struct {
	Concert
	VenueName  string
	ArtistName string
}

type Venue struct {
	ID      int
	Name    string
	City    string
	Country string
}

type Artist struct {
	ID   int
	Name string
}

type User struct {
	ID           int
	Email        string
	PasswordHash string
	SNS          []string
}

type PublicUser struct {
	ID    int
	Email string
}

func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email}
}

type Passkey struct {
	CredentialID   string
	PublicKey      string
	SignCount      int
	UserPresent    bool
	UserVerified   bool
	BackupEligible bool
	BackupState    bool
}

// PublicPasskey is the safe representation returned to the front-end.
type PublicPasskey struct {
	CredentialID string
	SignCount    int
}

func (p Passkey) Public() PublicPasskey {
	return PublicPasskey{CredentialID: p.CredentialID, SignCount: p.SignCount}
}

type WebAuthnChallengeRow struct {
	ID          int
	SessionData string
}

type WTType string

const (
	WTB WTType = "wtb"
	WTS WTType = "wts"
)

type WT struct {
	UserID    int
	ConcertID int
	Type      WTType
}

type ProfileWT struct {
	Type    string
	Concert DisplayConcert
}

type ProfileAlert struct {
	ID         int
	TargetType string
	TargetID   int
	TargetName string
}

type Favorite struct {
	UserID    int
	ConcertID int
}

type Alert struct {
	AlertID    int
	UserID     int
	TargetType string
	TargetID   int
}

type Setlist struct {
	ConcertID int
	Songs     []string
}

type SyncTicketmaster struct {
	LastPublicVisibilityStartDateTime time.Time
}

type DataSet struct {
	Venues    []Venue
	Artists   []Artist
	Concerts  []Concert
	Users     []User
	WTs       []WT
	Favorites []Favorite
	Alerts    []Alert
	Setlists  []Setlist
	Sync      SyncTicketmaster
}
