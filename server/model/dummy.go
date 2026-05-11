package model

import "time"

func DummyData() DataSet {
	venues := []Venue{
		{
			ID:      1,
			Name:    "Accor Arena",
			City:    "Paris",
			Country: "FR",
		},
		{
			ID:      2,
			Name:    "Uber Arena",
			City:    "Berlin",
			Country: "DE",
		},
		{
			ID:      3,
			Name:    "O2 Arena",
			City:    "London",
			Country: "GB",
		},
	}

	artists := []Artist{
		{ID: 1, Name: "NMIXX"},
		{ID: 2, Name: "Dreamcatcher"},
	}

	concerts := []Concert{
		{
			ID:                1,
			Name:              "NMIXX - Europe Tour",
			Date:              time.Date(2026, 3, 24, 20, 0, 0, 0, time.UTC),
			VenueID:           venues[0].ID,
			ArtistID:          artists[0].ID,
			URL:               "https://example.com/nmixx",
			Photos:            []string{"https://example.com/nmixx.jpg"},
			SaleStartDateTime: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:                2,
			Name:              "Dreamcatcher - Apocalypse Tour",
			Date:              time.Date(2026, 4, 2, 19, 30, 0, 0, time.UTC),
			VenueID:           venues[1].ID,
			ArtistID:          artists[1].ID,
			URL:               "https://example.com/dreamcatcher",
			Photos:            []string{"https://example.com/dreamcatcher.jpg"},
			SaleStartDateTime: time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:                3,
			Name:              "NMIXX - London",
			Date:              time.Date(2026, 4, 5, 20, 0, 0, 0, time.UTC),
			VenueID:           venues[2].ID,
			ArtistID:          artists[0].ID,
			URL:               "https://example.com/nmixx-london",
			Photos:            []string{"https://example.com/nmixx-london.jpg"},
			SaleStartDateTime: time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC),
		},
	}

	users := []User{
		{ID: 1, Pseudo: "alice", SNS: []string{"@alice"}},
		{ID: 2, Pseudo: "bob", SNS: []string{"@bob"}},
	}

	wts := []WT{
		{UserID: users[0].ID, ConcertID: concerts[0].ID, Type: WTB},
		{UserID: users[1].ID, ConcertID: concerts[1].ID, Type: WTS},
	}

	favoris := []Favoris{
		{UserID: users[0].ID, ConcertID: concerts[0].ID},
		{UserID: users[1].ID, ConcertID: concerts[2].ID},
	}

	alerts := []Alert{
		{AlertID: 1, UserID: users[0].ID, CibleType: "artist", CibleID: artists[0].ID},
		{AlertID: 2, UserID: users[1].ID, CibleType: "venue", CibleID: venues[1].ID},
	}

	setlists := []Setlist{
		{ConcertID: concerts[0].ID, Songs: []string{"Song A", "Song B"}},
		{ConcertID: concerts[1].ID, Songs: []string{"Intro", "Scream", "Silent Night"}},
		{ConcertID: concerts[2].ID, Songs: []string{"Track 1", "Track 2"}},
	}

	sync := SyncTicketmaster{
		LastPublicVisibilityStartDateTime: time.Now().UTC(),
	}

	return DataSet{
		Venues:   venues,
		Artists:  artists,
		Concerts: concerts,
		Users:    users,
		WTs:      wts,
		Favoris:  favoris,
		Alerts:   alerts,
		Setlists: setlists,
		Sync:     sync,
	}
}
