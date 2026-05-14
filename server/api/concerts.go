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
		SELECT *
		FROM concerts
		WHERE (? IS NULL OR artist_id = ?) AND (? IS NULL OR venue_id = ?)
		ORDER BY date`, model.ScanConcert, artistFilter, artistFilter, venueFilter, venueFilter)
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
		SELECT *
		FROM concerts
		WHERE id = ?`, model.ScanConcert, id)
}
