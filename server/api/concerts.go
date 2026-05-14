package api

import (
	"database/sql"
	"net/http"
	"server/model"
	"strconv"
)

// Get a list of concerts, optionally filtered by artistID and/or venueID
func handleConcerts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get optional artistID and venueID query parameters
	artistFilter := httpGetOptionalIntParam(r, "artistID")
	venueFilter := httpGetOptionalIntParam(r, "venueID")

	// Query concerts with optional filters
	sqlQueryList(w, r, db, "concerts", `
		SELECT c.id,
			c.name,
			c.date,
			c.venue_id,
			c.artist_id,
			c.url,
			c.photo_url,
			c.seatmap_url,
			c.sale_start_datetime,
			v.name,
			a.name
		FROM concerts c
		JOIN venues v ON v.id = c.venue_id
		JOIN artists a ON a.id = c.artist_id
		WHERE (? IS NULL OR c.artist_id = ?) AND (? IS NULL OR c.venue_id = ?)
		ORDER BY c.date`, model.ScanDisplayConcert, artistFilter, artistFilter, venueFilter, venueFilter)
}

// Get a concert by ID
func handleConcertByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the id parameter from the URL path
	idParam, ok := httpGetParam(w, r, "id")
	if !ok {
		return
	}

	// Convert id to integer
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return
	}

	// Query the concert by ID
	sqlQueryOne(w, r, db, "concert", `
		SELECT c.id,
			c.name,
			c.date,
			c.venue_id,
			c.artist_id,
			c.url,
			c.photo_url,
			c.seatmap_url,
			c.sale_start_datetime,
			v.name,
			a.name
		FROM concerts c
		JOIN venues v ON v.id = c.venue_id
		JOIN artists a ON a.id = c.artist_id
		WHERE c.id = ?`, model.ScanDisplayConcert, id)
}
