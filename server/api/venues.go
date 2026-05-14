package api

import (
	"database/sql"
	"net/http"

	"server/model"
)

// Get a list of venues, optionally filtered by search query
func handleVenues(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get optional search query parameter
	search := sqlLikeSearch(r.URL.Query().Get("search"))
	
	// Query venues with optional search filter
	sqlQueryList(w, r, db, "venues", `
		SELECT id, name, city, country
		FROM venues
		WHERE ? = '' OR lower(name) LIKE ?
		ORDER BY name`, model.ScanVenue, search, search)
}
