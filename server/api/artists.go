package api

import (
	"net/http"
	"strings"

	"server/model"
)

func handleArtists(w http.ResponseWriter, r *http.Request, data model.DataSet) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	
	results := make([]model.Artist, 0, len(data.Artists))
	for _, artist := range data.Artists {
		if query == "" || strings.Contains(strings.ToLower(artist.Name), query) {
			results = append(results, artist)
		}
	}

	writeJSON(w, results)
}
