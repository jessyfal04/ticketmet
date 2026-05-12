package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func ServeMux(db *sql.DB, clientDir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", withLogging(handleHealthz))
	mux.HandleFunc("/artists", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleArtists(w, r, db)
	}))
	mux.HandleFunc("/venues", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleVenues(w, r, db)
	}))
	mux.HandleFunc("/concerts", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleConcerts(w, r, db)
	}))
	mux.HandleFunc("/concerts/", withLogging(func(w http.ResponseWriter, r *http.Request) {
		handleConcertByID(w, r, db)
	}))
	mux.Handle("/", withLogging(http.FileServer(http.Dir(clientDir)).ServeHTTP))
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

// HELPERS HTTP

func httpInternalServerError(w http.ResponseWriter, message string, err error) {
	log.Printf("%s: %v", message, err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func httpNotFoundError(w http.ResponseWriter) {
	http.Error(w, "not found", http.StatusNotFound)
}

func httpMethodNotAllowedError(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func httpBadRequestError(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

// HELPERS ROUTE

func pathIntParam(w http.ResponseWriter, r *http.Request, prefix string, message string) (int, bool) {
	value := strings.TrimPrefix(r.URL.Path, prefix)
	if value == "" || strings.Contains(value, "/") {
		httpBadRequestError(w, message)
		return 0, false
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		httpBadRequestError(w, message)
		return 0, false
	}
	return id, true
}

// HELPERS SQL

func sqlLikeSearch(search string) string {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return ""
	}
	return "%" + search + "%"
}

func sqlOptionalInt(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	i, _ := strconv.Atoi(value)
	return i
}

func sqlQueryList[T any](w http.ResponseWriter, r *http.Request, db *sql.DB, label string, query string, scan func(interface{ Scan(...any) error }) (T, error), args ...any) {
	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		httpInternalServerError(w, label+" query failed", err)
		return
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			httpInternalServerError(w, label+" scan failed", err)
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		httpInternalServerError(w, label+" rows failed", err)
		return
	}

	writeJSON(w, results)
}

func sqlQueryOne[T any](w http.ResponseWriter, r *http.Request, db *sql.DB, label string, query string, scan func(interface{ Scan(...any) error }) (T, error), args ...any) {
	row := db.QueryRowContext(r.Context(), query, args...)
	item, err := scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		httpNotFoundError(w)
		return
	}
	if err != nil {
		httpInternalServerError(w, label+" query failed", err)
		return
	}

	writeJSON(w, item)
}
