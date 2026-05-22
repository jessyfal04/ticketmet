package api

import (
	"net/http"
	"server/job"
	"server/model"
)

// Get a list of artists, optionally filtered by search query
func handleArtists(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get optional search query parameter
		search := job.SqlLikeSearch(r.URL.Query().Get("search"))

		// Query artists with optional search filter
		artists, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT id, name
		FROM artists
		WHERE ? = '' OR lower(name) LIKE ?
		ORDER BY name`, model.ScanArtist, search, search)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, artists)
	}
}
