package api

import (
	"encoding/json"
	"log"
	"net/http"
	"server/job"
	"strconv"
	"strings"
)

// ROUTES

func ServeMux(clientDir string, dbChan chan<- job.DBRequest, mailChan chan<- job.Envelope, setlistChan chan<- job.SetlistRequest) *http.ServeMux {
	mux := http.NewServeMux()

	// health.go
	mux.HandleFunc("GET /healthz", toMuxHandler(handleHealthz()))

	// artists.go
	mux.HandleFunc("GET /api/artists", toMuxHandler(handleArtists(dbChan)))

	// venues.go
	mux.HandleFunc("GET /api/venues", toMuxHandler(handleVenues(dbChan)))

	// concerts.go
	mux.HandleFunc("GET /api/concerts", toMuxHandler(handleConcerts(dbChan)))
	mux.HandleFunc("GET /api/concerts/{id}", toMuxHandler(handleConcertByID(dbChan)))

	// setlist.go
	mux.HandleFunc("GET /api/setlist/{id}", toMuxHandler(handleConcertSetlist(setlistChan)))

	// favorites.go
	mux.HandleFunc("GET /api/favorites/{id}", toMuxHandler(handleConcertSNS(dbChan)))
	mux.HandleFunc("POST /api/favorites/{id}", toMuxHandler(handleFavoriteAdd(dbChan)))
	mux.HandleFunc("DELETE /api/favorites/{id}", toMuxHandler(handleFavoriteDelete(dbChan)))

	// wt.go
	mux.HandleFunc("GET /api/wt/{id}", toMuxHandler(handleConcertWT(dbChan)))
	mux.HandleFunc("POST /api/wt/{id}", toMuxHandler(handleWTAdd(dbChan)))
	mux.HandleFunc("DELETE /api/wt/{id}", toMuxHandler(handleWTDelete(dbChan)))

	// profile.go
	mux.HandleFunc("GET /api/me", toMuxHandler(handleProfileGet(dbChan)))
	mux.HandleFunc("PATCH /api/me", toMuxHandler(handleProfilePatch(dbChan)))

	// alerts.go
	mux.HandleFunc("POST /api/alerts", toMuxHandler(handleAlertCreate(dbChan)))
	mux.HandleFunc("DELETE /api/alerts/{alertId}", toMuxHandler(handleAlertDelete(dbChan)))

	// auth.go
	mux.HandleFunc("POST /api/auth/register", toMuxHandler(handleRegister(dbChan, mailChan)))
	mux.HandleFunc("POST /api/auth/login", toMuxHandler(handleLogin(dbChan)))
	mux.HandleFunc("POST /api/auth/logout", toMuxHandler(handleLogout(dbChan)))
	mux.HandleFunc("DELETE /api/auth/unregister", toMuxHandler(handleUnregister(dbChan)))
	mux.HandleFunc("GET /api/auth/me", toMuxHandler(handleMe(dbChan)))
	mux.HandleFunc("GET /api/auth/email-exists", toMuxHandler(handleEmailExists(dbChan)))

	// passkeys.go
	mux.HandleFunc("POST /api/auth/passkeys/register/begin", toMuxHandler(handlePasskeyRegisterBegin(dbChan)))
	mux.HandleFunc("POST /api/auth/passkeys/register/finish", toMuxHandler(handlePasskeyRegisterFinish(dbChan)))
	mux.HandleFunc("POST /api/auth/passkeys/login/begin", toMuxHandler(handlePasskeyLoginBegin(dbChan)))
	mux.HandleFunc("POST /api/auth/passkeys/login/finish", toMuxHandler(handlePasskeyLoginFinish(dbChan)))
	mux.HandleFunc("GET /api/auth/passkeys", toMuxHandler(handlePasskeyList(dbChan)))
	mux.HandleFunc("DELETE /api/auth/passkeys/{credentialId}", toMuxHandler(handlePasskeyDelete(dbChan)))

	// Frontend
	mux.Handle("GET /", toMuxHandler(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(clientDir)).ServeHTTP(w, r)
	}))
	return mux
}

func toMuxHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		handler(w, r)
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

func httpGetStringParam(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	value := strings.TrimSpace(r.PathValue(name))
	if value == "" {
		value = strings.TrimSpace(r.URL.Query().Get(name))
	}
	if value == "" {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return "", false
	}
	return value, true
}

func httpGetIntParam(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	value, ok := httpGetStringParam(w, r, name)
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return 0, false
	}
	return i, true
}

func httpGetOptionalIntParam(w http.ResponseWriter, r *http.Request, name string) (any, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return nil, true
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", nil)
		return nil, false
	}
	return i, true
}
