package api

import (
	"net/http"
	"server/job"
)

type setlistResponse struct {
	Songs   []string
	Fetched bool
}

// Get the potential setlist for a concert.
func handleConcertSetlist(setlistChan chan<- job.SetlistRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		concertID, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		result, err := job.RequestSetlist(r.Context(), setlistChan, concertID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, setlistResponse{Songs: result.Songs, Fetched: result.Fetched})
	}
}
