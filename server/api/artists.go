package api

import (
	"database/sql"
	"net/http"
	"server/model"
)

// Get a list of artists, optionally filtered by search query
func handleArtists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get optional search query parameter
	search := sqlLikeSearch(r.URL.Query().Get("search"))
	
	// Query artists with optional search filter
	sqlQueryList(w, r, db, "artists", `
		SELECT id, name
		FROM artists
		WHERE ? = '' OR lower(name) LIKE ?
		ORDER BY name`, model.ScanArtist, search, search)
}
