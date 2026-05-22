package api

import (
	"database/sql"
	"net/http"
	"strings"
)

type alertResponse struct {
	ID         int
	TargetType string
	TargetID   int
	TargetName string
}

type alertCreateResponse struct {
	OK bool
}

// Create an artist or venue alert.
func handleAlertCreate(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	targetType, ok := httpGetStringParam(w, r, "targetType")
	if !ok {
		return
	}
	targetType = strings.ToLower(targetType)
	if targetType != "artist" && targetType != "venue" {
		logHttpError(w, http.StatusBadRequest, "invalid target type", nil)
		return
	}

	targetID, ok := httpGetIntParam(w, r, "targetId")
	if !ok {
		return
	}
	if targetID <= 0 {
		logHttpError(w, http.StatusBadRequest, "invalid target id", nil)
		return
	}

	if !targetExists(w, r, db, targetType, targetID) {
		return
	}

	if err := sqlExec(r, db, `
		INSERT OR IGNORE INTO alerts (user_id, target_type, target_id)
		VALUES (?, ?, ?)`, user.ID, targetType, targetID); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	writeJSON(w, alertCreateResponse{OK: true})
}

// Delete one current-user alert.
func handleAlertDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	alertID, ok := httpGetIntParam(w, r, "alertId")
	if !ok {
		return
	}

	if err := sqlExec(r, db, `
		DELETE FROM alerts
		WHERE id = ? AND user_id = ?`, alertID, user.ID); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func targetExists(w http.ResponseWriter, r *http.Request, db *sql.DB, targetType string, targetID int) bool {
	query := "SELECT EXISTS(SELECT 1 FROM venues WHERE id = ?)"
	if targetType == "artist" {
		query = "SELECT EXISTS(SELECT 1 FROM artists WHERE id = ?)"
	}

	exists, err := sqlQueryBool(r, db, query, targetID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return false
	}
	if !exists {
		logHttpError(w, http.StatusNotFound, "target not found", nil)
		return false
	}
	return true
}
