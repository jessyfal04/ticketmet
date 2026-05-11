package api

import (
	"net/http"
	"strings"

	"server/model"
)

func handleVenues(w http.ResponseWriter, r *http.Request, data model.DataSet) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	
	results := make([]model.Venue, 0, len(data.Venues))
	for _, venue := range data.Venues {
		if query == "" || strings.Contains(strings.ToLower(venue.Name), query) {
			results = append(results, venue)
		}
	}

	writeJSON(w, results)
}
