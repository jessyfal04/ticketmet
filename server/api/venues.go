package api

import (
	"database/sql"
	"net/http"

	"server/model"
)

func handleVenues(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		httpMethodNotAllowedError(w)
		return
	}

	search := sqlLikeSearch(r.URL.Query().Get("search"))
	sqlQueryList(w, r, db, "venues", `
		SELECT id, name, city, country
		FROM venues
		WHERE ? = '' OR lower(name) LIKE ?
		ORDER BY name`, model.ScanVenue, search, search)
}
