package api

import (
	"database/sql"
	"net/http"
	"server/model"
	"strconv"
	"strings"
)

type wtResponse struct {
	WTB []string
	WTS []string
}

// Get WTB/WTS information for a concert.
func handleConcertWT(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get concert ID
	idParam, ok := httpGetParam(w, r, "id")
	if !ok {
		return
	}
	concertID, err := strconv.Atoi(idParam)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return
	}

	// Query buyers
	wtb, err := sqlScanList(r, db, `
		SELECT DISTINCT us.sns
		FROM wt
		JOIN user_sns us ON us.user_id = wt.user_id
		WHERE wt.concert_id = ? AND wt.type = 'wtb'
		ORDER BY us.sns`, model.ScanString, concertID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	// Query sellers
	wts, err := sqlScanList(r, db, `
		SELECT DISTINCT us.sns
		FROM wt
		JOIN user_sns us ON us.user_id = wt.user_id
		WHERE wt.concert_id = ? AND wt.type = 'wts'
		ORDER BY us.sns`, model.ScanString, concertID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	writeJSON(w, wtResponse{WTB: wtb, WTS: wts})
}

// Add the current user to WTB or WTS for a concert.
func handleWTAdd(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Get concert ID
	idParam, ok := httpGetParam(w, r, "id")
	if !ok {
		return
	}
	concertID, err := strconv.Atoi(idParam)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return
	}

	// Get WTB/WTS type
	wtType, ok := wtTypeParam(w, r)
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

	// Insert WTB/WTS row
	if err := sqlExec(r, db, `
		INSERT OR IGNORE INTO wt (user_id, concert_id, type)
		VALUES (?, ?, ?)`, user.ID, concertID, wtType); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Remove the current user from WTB or WTS for a concert.
func handleWTDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Get concert ID
	idParam, ok := httpGetParam(w, r, "id")
	if !ok {
		return
	}
	concertID, err := strconv.Atoi(idParam)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return
	}

	// Get WTB/WTS type
	wtType, ok := wtTypeParam(w, r)
	if !ok {
		return
	}

	// Delete WTB/WTS row
	if err := sqlExec(r, db, `
		DELETE FROM wt
		WHERE user_id = ? AND concert_id = ? AND type = ?`, user.ID, concertID, wtType); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func wtTypeParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	value, ok := httpGetParam(w, r, "type")
	if !ok {
		return "", false
	}
	wtType := strings.ToLower(value)
	if wtType != "wtb" && wtType != "wts" {
		logHttpError(w, http.StatusBadRequest, "invalid wt type", nil)
		return "", false
	}
	return wtType, true
}
