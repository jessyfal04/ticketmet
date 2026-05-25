package api

import (
	"net/http"
	"server/job"
	"server/model"
	"strings"
)

type alertCreateResponse struct {
	OK bool
}

// Create an artist or venue alert.
func handleAlertCreate(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireUser(w, r, dbChan)
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

		query := "SELECT EXISTS(SELECT 1 FROM venues WHERE id = ?)"
		if targetType == "artist" {
			query = "SELECT EXISTS(SELECT 1 FROM artists WHERE id = ?)"
		}
		exists, err := job.SqlScanOne(r.Context(), dbChan, query, model.ScanBool, targetID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		if !exists {
			logHttpError(w, http.StatusNotFound, "target not found", nil)
			return
		}
		if err := job.SqlExec(r.Context(), dbChan, `
		INSERT OR IGNORE INTO alerts (user_id, target_type, target_id)
		VALUES (?, ?, ?)`, user.ID, targetType, targetID); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		writeJSON(w, alertCreateResponse{OK: true})
	}
}

// Delete one current-user alert.
func handleAlertDelete(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		alertID, ok := httpGetIntParam(w, r, "alertId")
		if !ok {
			return
		}

		if err := job.SqlExec(r.Context(), dbChan, `
		DELETE FROM alerts
		WHERE id = ? AND user_id = ?`, alertID, user.ID); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
