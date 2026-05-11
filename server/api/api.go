package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"server/model"
)

func ServeMux(data model.DataSet) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/artists", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleArtists(w, r, data)
	}))
	mux.HandleFunc("/artistes", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleArtists(w, r, data)
	}))
	mux.HandleFunc("/venues", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleVenues(w, r, data)
	}))
	mux.HandleFunc("/salles", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleVenues(w, r, data)
	}))
	mux.HandleFunc("/concerts", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleConcerts(w, r, data)
	}))
	mux.HandleFunc("/concerts/", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleConcertByID(w, r, data)
	}))
	mux.Handle("/", withLogging(http.FileServer(http.Dir("../client")).ServeHTTP))
	return mux
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		next(w, r)
	}
}

func parseOptionalInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	i, _ := strconv.Atoi(s)
	return &i
}
