package api

import (
	"database/sql"
	"net/http"
)

// Get a simple health check response
func handleHealthz(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
