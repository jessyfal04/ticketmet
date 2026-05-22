package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/mail"
	"server/job"
	"server/model"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authUserResponse struct {
	User model.PublicUser
}

type emailExistsResponse struct {
	Exists bool
}

// Register a new user and open a session
func handleRegister(mailChan chan<- job.Envelope) func(http.ResponseWriter, *http.Request, *sql.DB) {
	return func(w http.ResponseWriter, r *http.Request, db *sql.DB) {
		// Get email and password from request body
		var body struct {
			Email    string
			Password string
		}
		if !readJSON(w, r, &body) {
			return
		}

		// Validate email and password
		email := strings.ToLower(strings.TrimSpace(body.Email))
		if _, err := mail.ParseAddress(email); err != nil || !strings.Contains(email, "@") {
			logHttpError(w, http.StatusBadRequest, "invalid email", err)
			return
		}
		if len(body.Password) < 8 {
			logHttpError(w, http.StatusBadRequest, "password must be at least 8 characters", nil)
			return
		}

		// Check email availability
		exists, err := sqlQueryBool(r, db, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		if exists {
			logHttpError(w, http.StatusConflict, "email already exists", nil)
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		// Insert user
		err = sqlExec(r, db, `
				INSERT INTO users (email, password_hash)
				VALUES (?, ?)`, email, string(hash))
		if err != nil {
			logHttpError(w, http.StatusConflict, "email already exists", nil)
			return
		}

		// Reload inserted user
		user, err := sqlScanOne(r, db, `
				SELECT id, email, password_hash
				FROM users
				WHERE email = ?`, model.ScanUser, email)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		// Open session
		if !createSessionCookie(w, r, db, user.ID) {
			return
		}

		// Send welcome email
		envelope := job.Envelope{
			Dst:     user.Email,
			Message: job.WelcomeMail(user.Email),
		}
		mailChan <- envelope

		writeJSON(w, authUserResponse{User: user.Public()})
	}
}

// Check password and open a session
func handleLogin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get email and password from request body
	var body struct {
		Email    string
		Password string
	}
	if !readJSON(w, r, &body) {
		return
	}

	// Check credentials
	user, err := sqlScanOne(r, db, `
			SELECT id, email, password_hash
			FROM users
			WHERE email = ?`, model.ScanUser, body.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) != nil {
		logHttpError(w, http.StatusUnauthorized, "invalid email or password", nil)
		return
	}

	// Open session with a cookie
	if !createSessionCookie(w, r, db, user.ID) {
		return
	}
	writeJSON(w, authUserResponse{User: user.Public()})
}

// Delete the current session
func handleLogout(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the cookie and delete the session from database
	if cookie, err := r.Cookie("session"); err == nil {
		if err := sqlExec(r, db, "DELETE FROM sessions WHERE token_hash = ?", hashToken(cookie.Value)); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
	}

	http.SetCookie(w, sessionCookie(r, "", -1))
	w.WriteHeader(http.StatusNoContent)
}

// Check password and delete the user
func handleUnregister(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Get password from request body
	var body struct {
		Password string
	}
	if !readJSON(w, r, &body) {
		return
	}

	// Check password
	if user.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) != nil {
		logHttpError(w, http.StatusUnauthorized, "invalid password", nil)
		return
	}

	// Delete sessions then user
	err := sqlExec(r, db, "DELETE FROM sessions WHERE user_id = ?", user.ID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	err = sqlExec(r, db, "DELETE FROM users WHERE id = ?", user.ID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	http.SetCookie(w, sessionCookie(r, "", -1))
	w.WriteHeader(http.StatusNoContent)
}

// Return the current user
func handleMe(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}
	writeJSON(w, authUserResponse{User: user.Public()})
}

// Check if an email is already used
func handleEmailExists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get email from query parameter
	value, ok := httpGetStringParam(w, r, "email")
	if !ok {
		return
	}

	// Format email
	email := strings.ToLower(value)

	// Query user existence
	exists, err := sqlQueryBool(r, db, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	writeJSON(w, emailExistsResponse{Exists: exists})
}

// Get the current user from the session cookie
func requireUser(w http.ResponseWriter, r *http.Request, db *sql.DB) (model.User, bool) {
	// Get session cookie
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		logHttpError(w, http.StatusUnauthorized, "authentication required", nil)
		return model.User{}, false
	}

	// Load user by session token
	user, err := sqlScanOne(r, db, `
			SELECT u.id, u.email, u.password_hash
			FROM sessions s
			JOIN users u ON u.id = s.user_id
			WHERE s.token_hash = ?
				AND s.expires_at > strftime('%s','now')`,
		model.ScanUser, hashToken(cookie.Value))

	if err != nil {
		logHttpError(w, http.StatusUnauthorized, "authentication required", nil)
		return model.User{}, false
	}
	return user, true
}

// createSessionCookie stores a server session and sends its cookie.
func createSessionCookie(w http.ResponseWriter, r *http.Request, db *sql.DB, userID int) bool {
	// Generate a random token that expires in 10 minutes (1 week in comment)
	token := rand.Text()
	expires := time.Now().UTC().Add(10 * time.Minute) // time.Now().UTC().Add(7 * 24 * time.Hour)

	// Delete old token
	err := sqlExec(r, db, "DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return false
	}

	// Store new token hash in database
	err = sqlExec(r, db, "INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)",
		userID, hashToken(token), expires.Unix())

	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return false
	}

	http.SetCookie(w, sessionCookie(r, token, int(time.Until(expires).Seconds())))
	return true
}

// Build the session cookie with appropriate flags
func sessionCookie(r *http.Request, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "session",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}

func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// hashToken hashes a token before database storage
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
