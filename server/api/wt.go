package api

import (
	"net/http"
	"server/job"
	"server/model"
	"strings"
)

type wtResponse struct {
	WTB []string
	WTS []string
}

// Get WTB/WTS information for a concert.
func handleConcertWT(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get concert ID
		concertID, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		// Query buyers
		wtb, err := job.SqlScanList(r.Context(), dbChan, `
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
		wts, err := job.SqlScanList(r.Context(), dbChan, `
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
}

// Add the current user to WTB or WTS for a concert.
func handleWTAdd(dbChan chan<- job.DBRequest) http.HandlerFunc {
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

		// Get WTB/WTS type
		wtType, ok := wtTypeParam(w, r)
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

		// Insert WTB/WTS row
		if err := job.SqlExec(r.Context(), dbChan, `
		INSERT OR IGNORE INTO wt (user_id, concert_id, type)
		VALUES (?, ?, ?)`, user.ID, concertID, wtType); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Remove the current user from WTB or WTS for a concert.
func handleWTDelete(dbChan chan<- job.DBRequest) http.HandlerFunc {
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

		// Get WTB/WTS type
		wtType, ok := wtTypeParam(w, r)
		if !ok {
			return
		}

		// Delete WTB/WTS row
		if err := job.SqlExec(r.Context(), dbChan, `
		DELETE FROM wt
		WHERE user_id = ? AND concert_id = ? AND type = ?`, user.ID, concertID, wtType); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func wtTypeParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	value, ok := httpGetStringParam(w, r, "type")
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
