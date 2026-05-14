package api

import (
	"database/sql"
	"net/http"
	"server/model"
)

func handleConcerts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		httpMethodNotAllowedError(w)
		return
	}

	artistIDParam := r.URL.Query().Get("artistID")
	venueIDParam := r.URL.Query().Get("venueID")

	artistFilter := sqlOptionalInt(artistIDParam)
	venueFilter := sqlOptionalInt(venueIDParam)

	sqlQueryList(w, r, db, "concerts", `
		SELECT id, name, date, venue_id, artist_id, url, photo_url, seatmap_url, sale_start_datetime
		FROM concerts
		WHERE (? IS NULL OR artist_id = ?) AND (? IS NULL OR venue_id = ?)
		ORDER BY date`, model.ScanConcert, artistFilter, artistFilter, venueFilter, venueFilter)
}

func handleConcertByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		httpMethodNotAllowedError(w)
		return
	}

	id, ok := pathIntParam(w, r, "/concerts/", "invalid concertID")
	if !ok {
		return
	}

	sqlQueryOne(w, r, db, "concert", `
		SELECT id, name, date, venue_id, artist_id, url, photo_url, seatmap_url, sale_start_datetime
		FROM concerts
		WHERE id = ?`, model.ScanConcert, id)
}
