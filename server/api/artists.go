package api

import (
	"database/sql"
	"net/http"
	"server/model"
)

func handleArtists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		httpMethodNotAllowedError(w)
		return
	}

	search := sqlLikeSearch(r.URL.Query().Get("search"))
	sqlQueryList(w, r, db, "artists", `
		SELECT id, name
		FROM artists
		WHERE ? = '' OR lower(name) LIKE ?
		ORDER BY name`, model.ScanArtist, search, search)
}
