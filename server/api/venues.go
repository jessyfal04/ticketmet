package api

import (
	"net/http"
	"server/job"
	"server/model"
)

// Get a list of venues, optionally filtered by search query
func handleVenues(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get optional search query parameter
		search := job.SqlLikeSearch(r.URL.Query().Get("search"))

		// Query venues with optional search filter
		venues, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT id, name, city, country
		FROM venues
		WHERE ? = '' OR lower(name) LIKE ?
		ORDER BY name`, model.ScanVenue, search, search)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, venues)
	}
}
