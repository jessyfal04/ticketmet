package api

import (
	"database/sql"
	"net/http"
	"server/model"
	"strconv"
)

type setlistResponse struct {
	Songs []string
}

// Get the potential setlist for a concert.
func handleConcertSetlist(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	// Query setlist songs
	songs, err := sqlScanList(r, db, `
		SELECT song_name
		FROM setlists
		WHERE concert_id = ?
		ORDER BY song_order`, model.ScanString, concertID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	writeJSON(w, setlistResponse{Songs: songs})
}
