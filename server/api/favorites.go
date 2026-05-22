package api

import (
	"database/sql"
	"net/http"
	"server/model"
)

type concertSNSResponse struct {
	SNS []string
}

// Get public SNS handles from users interested in the same concert.
func handleConcertSNS(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get concert ID
	concertID, ok := httpGetIntParam(w, r, "id")
	if !ok {
		return
	}

	// Query SNS handles from users who favorited this concert
	sns, err := sqlScanList(r, db, `
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

// Add the current concert to the user's favorites.
func handleFavoriteAdd(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Get concert ID
	concertID, ok := httpGetIntParam(w, r, "id")
	if !ok {
		return
	}

	// Check concert exists
	exists, err := sqlQueryBool(r, db, "SELECT EXISTS(SELECT 1 FROM concerts WHERE id = ?)", concertID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	if !exists {
		logHttpError(w, http.StatusNotFound, "concert not found", nil)
		return
	}

	// Insert favorite
	if err := sqlExec(r, db, `
		INSERT OR IGNORE INTO favorites (user_id, concert_id)
		VALUES (?, ?)`, user.ID, concertID); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Remove the current concert from the user's favorites.
func handleFavoriteDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Get concert ID
	concertID, ok := httpGetIntParam(w, r, "id")
	if !ok {
		return
	}

	// Delete favorite
	if err := sqlExec(r, db, `
		DELETE FROM favorites
		WHERE user_id = ? AND concert_id = ?`, user.ID, concertID); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
