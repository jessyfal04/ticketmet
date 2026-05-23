package api

import (
	"database/sql"
	"net/http"
	"server/job"
	"server/model"
	"time"
)

const concertsPageSize = 20

// Get a list of concerts, optionally filtered by artistID and/or venueID
func handleConcerts(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get optional artistID and venueID query parameters
		artistFilter, ok := httpGetOptionalIntParam(w, r, "artistID")
		if !ok {
			return
		}
		venueFilter, ok := httpGetOptionalIntParam(w, r, "venueID")
		if !ok {
			return
		}
		countryFilter, ok := httpGetStringParam(w, r, "country")
		if !ok {
			return
		}
		statusFilter, ok := httpGetStringParam(w, r, "status")
		if !ok {
			return
		}
		page, ok := httpGetIntParam(w, r, "page")
		if !ok {
			return
		}
		if page < 1 {
			logHttpError(w, http.StatusBadRequest, "invalid page", nil)
			return
		}
		if statusFilter != "all" && statusFilter != "future" {
			logHttpError(w, http.StatusBadRequest, "invalid status", nil)
			return
		}
		offset := (page - 1) * concertsPageSize

		// Query concerts with optional filters
		concerts, err := job.SqlScanList(r.Context(), dbChan, `
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
		WHERE (? IS NULL OR c.artist_id = ?)
			AND (? IS NULL OR c.venue_id = ?)
			AND (? = 'all' OR v.country = ?)
			AND (? = 'all' OR c.date >= ?)
		ORDER BY c.date
		LIMIT ? OFFSET ?`, model.ScanDisplayConcert,
			artistFilter, artistFilter,
			venueFilter, venueFilter,
			countryFilter, countryFilter,
			statusFilter, time.Now().UTC().Format(time.RFC3339),
			concertsPageSize,
			offset)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, concerts)
	}
}

// Get a concert by ID
func handleConcertByID(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the id parameter from the URL path
		id, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		// Query the concert by ID
		concert, err := job.SqlScanOne(r.Context(), dbChan, `
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
		if err == sql.ErrNoRows {
			logHttpError(w, http.StatusNotFound, "", nil)
			return
		}
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, concert)
	}
}
