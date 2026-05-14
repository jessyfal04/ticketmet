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

// ROUTES

func ServeMux(db *sql.DB, clientDir string) *http.ServeMux {
	mux := http.NewServeMux()

	// health.go
	mux.HandleFunc("GET /healthz", route(db, handleHealthz))

	// artists.go
	mux.HandleFunc("GET /api/artists", route(db, handleArtists))

	// venues.go
	mux.HandleFunc("GET /api/venues", route(db, handleVenues))

	// concerts.go
	mux.HandleFunc("GET /api/concerts", route(db, handleConcerts))
	mux.HandleFunc("GET /api/concerts/{id}", route(db, handleConcertByID))

	// auth.go
	mux.HandleFunc("POST /api/auth/register", route(db, handleRegister))
	mux.HandleFunc("POST /api/auth/login", route(db, handleLogin))
	mux.HandleFunc("POST /api/auth/logout", route(db, handleLogout))
	mux.HandleFunc("DELETE /api/auth/unregister", route(db, handleUnregister))
	mux.HandleFunc("GET /api/auth/me", route(db, handleMe))
	mux.HandleFunc("GET /api/auth/email-exists", route(db, handleEmailExists))

	// passkeys.go
	mux.HandleFunc("POST /api/auth/passkeys/register/begin", route(db, handlePasskeyRegisterBegin))
	mux.HandleFunc("POST /api/auth/passkeys/register/finish", route(db, handlePasskeyRegisterFinish))
	mux.HandleFunc("POST /api/auth/passkeys/login/begin", route(db, handlePasskeyLoginBegin))
	mux.HandleFunc("POST /api/auth/passkeys/login/finish", route(db, handlePasskeyLoginFinish))
	mux.HandleFunc("GET /api/auth/passkeys", route(db, handlePasskeyList))
	mux.HandleFunc("DELETE /api/auth/passkeys/{credentialId}", route(db, handlePasskeyDelete))

	// Frontend
	mux.Handle("GET /", route(db, func(w http.ResponseWriter, r *http.Request, db *sql.DB) {
		http.FileServer(http.Dir(clientDir)).ServeHTTP(w, r)
	}))
	return mux
}

func route(db *sql.DB, handler func(http.ResponseWriter, *http.Request, *sql.DB)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		url := r.URL.String()
		log.Printf("%s %s", method, url)
		handler(w, r, db)
	}
}

// HELPERS JSON

// Write JSON response with the appropriate content type header
func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}

// readJSON reads JSON from the request body into dst
func readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		logHttpError(w, http.StatusBadRequest, "", err)
		return false
	}
	return true
}

// HELPERS HTTP
func logHttpError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Print(err)
	}
	if message == "" {
		message = http.StatusText(status)
	}
	http.Error(w, message, status)
}

func httpGetParam(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	value := r.PathValue(name)
	if value == "" {
		value = strings.TrimSpace(r.URL.Query().Get(name))
	}
	if value == "" {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return "", false
	}
	return value, true
}

func httpGetOptionalIntParam(r *http.Request, name string) any {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return nil
	}
	i, _ := strconv.Atoi(value)
	return i
}

// HELPERS SQL

func sqlLikeSearch(search string) string {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return ""
	}
	return "%" + search + "%"
}

func sqlQueryList[T any](w http.ResponseWriter, r *http.Request, db *sql.DB, label string, query string, scan func(interface{ Scan(...any) error }) (T, error), args ...any) {
	results, err := sqlScanList(r, db, query, scan, args...)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	writeJSON(w, results)
}

func sqlScanList[T any](r *http.Request, db *sql.DB, query string, scan func(interface{ Scan(...any) error }) (T, error), args ...any) ([]T, error) {
	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func sqlQueryOne[T any](w http.ResponseWriter, r *http.Request, db *sql.DB, label string, query string, scan func(interface{ Scan(...any) error }) (T, error), args ...any) {
	item, err := sqlScanOne(r, db, query, scan, args...)
	if errors.Is(err, sql.ErrNoRows) {
		logHttpError(w, http.StatusNotFound, "", nil)
		return
	}
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	writeJSON(w, item)
}

func sqlScanOne[T any](r *http.Request, db *sql.DB, query string, scan func(interface{ Scan(...any) error }) (T, error), args ...any) (T, error) {
	row := db.QueryRowContext(r.Context(), query, args...)
	return scan(row)
}

func sqlExec(r *http.Request, db *sql.DB, query string, args ...any) error {
	_, err := db.ExecContext(r.Context(), query, args...)
	return err
}

func sqlQueryBool(r *http.Request, db *sql.DB, query string, args ...any) (bool, error) {
	var value bool
	err := db.QueryRowContext(r.Context(), query, args...).Scan(&value)
	return value, err
}
