package api

import (
	"net/http"
	"strconv"
	"strings"

	"server/model"
)

func handleConcerts(w http.ResponseWriter, r *http.Request, data model.DataSet) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	artistIDParam := r.URL.Query().Get("artistID")
	if artistIDParam == "" {
		artistIDParam = r.URL.Query().Get("artisteId")
	}
	venueIDParam := r.URL.Query().Get("venueID")
	if venueIDParam == "" {
		venueIDParam = r.URL.Query().Get("salleId")
	}

	artistID := parseOptionalInt(artistIDParam)
	venueID := parseOptionalInt(venueIDParam)

	results := make([]model.Concert, 0, len(data.Concerts))
	for _, concert := range data.Concerts {
		if artistID == nil || *artistID == concert.ArtistID {
			if venueID == nil || *venueID == concert.VenueID {
				results = append(results, concert)
			}
		}
	}

	writeJSON(w, results)
}

func handleConcertByID(w http.ResponseWriter, r *http.Request, data model.DataSet) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/concerts/")
	if idStr == "" || strings.Contains(idStr, "/") {
		http.Error(w, "invalid concertID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid concertID", http.StatusBadRequest)
		return
	}

	for _, concert := range data.Concerts {
		if concert.ID == id {
			writeJSON(w, concert)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}
