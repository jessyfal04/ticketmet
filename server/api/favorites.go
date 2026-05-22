package api

import (
	"net/http"
	"server/job"
	"server/model"
)

type concertSNSResponse struct {
	SNS []string
}

// Get public SNS handles from users interested in the same concert.
func handleConcertSNS(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get concert ID
		concertID, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		// Query SNS handles from users who favorited this concert
		sns, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT DISTINCT us.sns
		FROM favorites f
		JOIN user_sns us ON us.user_id = f.user_id
		WHERE f.concert_id = ?
		ORDER BY us.sns`, model.ScanString, concertID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		writeJSON(w, concertSNSResponse{SNS: sns})
	}
}

// Add the current concert to the user's favorites.
func handleFavoriteAdd(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load current user
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		// Get concert ID
		concertID, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		// Check concert exists
		exists, err := job.SqlScanOne(r.Context(), dbChan, "SELECT EXISTS(SELECT 1 FROM concerts WHERE id = ?)", model.ScanBool, concertID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		if !exists {
			logHttpError(w, http.StatusNotFound, "concert not found", nil)
			return
		}

		// Insert favorite
		if err := job.SqlExec(r.Context(), dbChan, `
		INSERT OR IGNORE INTO favorites (user_id, concert_id)
		VALUES (?, ?)`, user.ID, concertID); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Remove the current concert from the user's favorites.
func handleFavoriteDelete(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load current user
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		// Get concert ID
		concertID, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		// Delete favorite
		if err := job.SqlExec(r.Context(), dbChan, `
		DELETE FROM favorites
		WHERE user_id = ? AND concert_id = ?`, user.ID, concertID); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
