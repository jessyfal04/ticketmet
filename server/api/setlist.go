package api

import (
	"net/http"
	"server/job"
)

type setlistResponse struct {
	Songs []string
}

// Get the potential setlist for a concert.
func handleConcertSetlist(setlistChan chan<- job.SetlistRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		concertID, ok := httpGetIntParam(w, r, "id")
		if !ok {
			return
		}

		ret := make(chan job.SetlistResult)
		setlistChan <- job.SetlistRequest{ConcertID: concertID, Ret: ret}
		result := <-ret
		if result.Err != nil {
			logHttpError(w, http.StatusInternalServerError, "", result.Err)
			return
		}
		writeJSON(w, setlistResponse{Songs: result.Songs})
	}
}
