package model

import "time"

type Concert struct {
	ID int
	Name string
	Date time.Time
	VenueID int
	ArtistID int
	URL string
	Photos []string
	SaleStartDateTime time.Time
}

type Venue struct {
	ID int
	Name string
	City string
	Country string
}

type Artist struct {
	ID int
	Name string
}

type User struct {
	ID int
	Pseudo string
	SNS []string
}

type WTType string

const (
	WTB WTType = "wtb"
	WTS WTType = "wts"
)

type WT struct {
	UserID int
	ConcertID int
	Type WTType
}

type Favoris struct {
	UserID int
	ConcertID int
}

type Alert struct {
	AlertID int
	UserID int
	CibleType string
	CibleID int
}

type Setlist struct {
	ConcertID int
	Songs []string
}

type SyncTicketmaster struct {
	LastPublicVisibilityStartDateTime time.Time
}

type DataSet struct {
	Venues []Venue
	Artists []Artist
	Concerts []Concert
	Users []User
	WTs []WT
	Favoris []Favoris
	Alerts []Alert
	Setlists []Setlist
	Sync SyncTicketmaster
}
