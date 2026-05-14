package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"server/model"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Model user.Model x webauthn.User
type webAuthnUser struct {
	model.User
	credentials []webauthn.Credential
}

// WebAuthnID returns the stable WebAuthn user handle.
func (u webAuthnUser) WebAuthnID() []byte {
	return []byte(strconv.Itoa(u.ID))
}

// WebAuthnName returns the user name shown to authenticators.
func (u webAuthnUser) WebAuthnName() string {
	return u.Email
}

// WebAuthnDisplayName returns the display name shown to authenticators.
func (u webAuthnUser) WebAuthnDisplayName() string {
	return u.Email
}

// WebAuthnCredentials returns the user's registered credentials.
func (u webAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// Start passkey creation for the current user
func handlePasskeyRegisterBegin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Build registration options
	wuser, err := loadWebAuthnUser(r, db, user.ID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	wa, err := newWebAuthn()
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	// Require resident key for passkey registration, so the user doesn't need to enter email during login
	options, wa_sessionData, err := wa.BeginRegistration(
		wuser,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
	)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", err)
		return
	}

	// Store the WebAuthn session server-side. The cookie only stores a random lookup token.
	if !saveWebAuthnChallenge(w, r, db, &user.ID, "registration", wa_sessionData) {
		return
	}
	writeJSON(w, options)
}

// Store the created passkey after browser verification
func handlePasskeyRegisterFinish(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Load the WebAuthn user
	wuser, err := loadWebAuthnUser(r, db, user.ID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	// Load the WebAuthn wa_sessionData saved during begin
	wa_sessionData, sessionID, err := loadWebAuthnChallenge(r, db, &user.ID, "registration")
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "passkey registration challenge not found", err)
		return
	}
	wa, err := newWebAuthn()
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	// Verify and insert credential
	credential, err := wa.FinishRegistration(wuser, wa_sessionData, r)
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "invalid passkey registration", err)
		return
	}

	// Insert the new credential into database
	err = sqlExec(r, db, `
		INSERT INTO webauthn_credentials (
			user_id, credential_id, public_key, sign_count
		)
		VALUES (?, ?, ?, ?)`,
		user.ID,
		base64.RawURLEncoding.EncodeToString(credential.ID),
		base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		credential.Authenticator.SignCount)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	// Delete the WebAuthn session
	err = sqlExec(r, db, "DELETE FROM webauthn_challenges WHERE id = ?", sessionID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	http.SetCookie(w, webAuthnChallengeCookie("", -1))

	w.WriteHeader(http.StatusCreated)
}

// Start passkey login
func handlePasskeyLoginBegin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Build login options, no need email
	wa, err := newWebAuthn()
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	// Build login options
	options, wa_sessionData, err := wa.BeginDiscoverableLogin()
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "", err)
		return
	}

	// Store the WebAuthn wa_sessionData server-side. The cookie only stores a random lookup token.
	if !saveWebAuthnChallenge(w, r, db, nil, "login", wa_sessionData) {
		return
	}
	writeJSON(w, options)
}

// Finish the login with the passkey
func handlePasskeyLoginFinish(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load the WebAuthn wa_sessionData saved during begin
	wa_sessionData, sessionID, err := loadWebAuthnChallenge(r, db, nil, "login")
	if err != nil {
		logHttpError(w, http.StatusBadRequest, "passkey login challenge not found", err)
		return
	}

	// Build WebAuthn config
	wa, err := newWebAuthn()
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	var user webAuthnUser
	var credential *webauthn.Credential

	// Verify credential
	resolvedUser, credential, err := wa.FinishPasskeyLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			id, err := strconv.Atoi(string(userHandle))
			if err != nil {
				return nil, err
			}
			return loadWebAuthnUser(r, db, id)
		}, wa_sessionData, r)
	if err != nil {
		logHttpError(w, http.StatusUnauthorized, "invalid passkey", err)
		return
	}
	// Type assert to our webAuthnUser model
	resolved, ok := resolvedUser.(webAuthnUser)
	if !ok {
		logHttpError(w, http.StatusUnauthorized, "invalid passkey user", nil)
		return
	}
	user = resolved

	// Update counter and remove the one-time WebAuthn wa_sessionData.
	err = sqlExec(r, db, `
		UPDATE webauthn_credentials
		SET sign_count = ?
		WHERE credential_id = ?`,
		credential.Authenticator.SignCount,
		base64.RawURLEncoding.EncodeToString(credential.ID))
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	// Delete the WebAuthn session
	err = sqlExec(r, db, "DELETE FROM webauthn_challenges WHERE id = ?", sessionID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	http.SetCookie(w, webAuthnChallengeCookie("", -1))
	if !createSessionCookie(w, r, db, user.ID) {
		return
	}
	writeJSON(w, map[string]model.PublicUser{"user": user.Public()})
}

// List current-user passkeys
func handlePasskeyList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Query credentials
	items, err := sqlScanList(r, db, `
		SELECT credential_id, public_key, sign_count
		FROM webauthn_credentials
		WHERE user_id = ?
		ORDER BY id`, model.ScanPublicPasskey, user.ID)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	writeJSON(w, map[string][]model.PublicPasskey{"passkeys": items})
}

// Delete one current-user passkey
func handlePasskeyDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Load current user
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	// Delete credential
	credentialId, ok := httpGetParam(w, r, "credentialId")
	if !ok {
		return
	}

	err := sqlExec(r, db, "DELETE FROM webauthn_credentials WHERE user_id = ? AND credential_id = ?", user.ID, credentialId)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper

// newWebAuthn builds WebAuthn config for the current host.
func newWebAuthn() (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPID:          "ticketmet.jessyfal04.dev",
		RPDisplayName: "TicketMet",
		RPOrigins:     []string{"https://ticketmet.jessyfal04.dev"},
	})
}

// Load a user with its credentials
func loadWebAuthnUser(r *http.Request, db *sql.DB, id int) (webAuthnUser, error) {
	user, err := sqlScanOne(r, db, `
		SELECT id, email, password_hash
		FROM users
		WHERE id = ?`, model.ScanUser, id)
	if err != nil {
		return webAuthnUser{}, err
	}

	passkeys, err := sqlScanList(r, db, `
		SELECT credential_id, public_key, sign_count
		FROM webauthn_credentials
		WHERE user_id = ?
		ORDER BY id`, model.ScanPasskey, id)
	if err != nil {
		return webAuthnUser{}, err
	}

	credentials := make([]webauthn.Credential, 0, len(passkeys))
	for _, passkey := range passkeys {
		id, err := base64.RawURLEncoding.DecodeString(passkey.CredentialID)
		if err != nil {
			return webAuthnUser{}, err
		}
		publicKey, err := base64.RawURLEncoding.DecodeString(passkey.PublicKey)
		if err != nil {
			return webAuthnUser{}, err
		}
		credentials = append(credentials, webauthn.Credential{
			ID:            id,
			PublicKey:     publicKey,
			Authenticator: webauthn.Authenticator{SignCount: uint32(passkey.SignCount)},
		})
	}
	return webAuthnUser{
		User:        user,
		credentials: credentials,
	}, nil
}

// Stores the WebAuthn SessionData
func saveWebAuthnChallenge(w http.ResponseWriter, r *http.Request, db *sql.DB, userID *int, kind string, wa_sessionData *webauthn.SessionData) bool {
	// Generate a random token for cookie storage
	token := rand.Text()

	// Serialize session data to JSON for database storage
	data, err := json.Marshal(wa_sessionData)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return false
	}

	// Store the session in database with an expiration time (10 minutes)
	expires := time.Now().UTC().Add(10 * time.Minute) // time.Now().UTC().Add(7 * 24 * time.Hour) for longer expiration in comment
	err = sqlExec(r, db, `
		INSERT INTO webauthn_challenges (user_id, token_hash, kind, session_data, expires_at)
		VALUES (?, ?, ?, ?, ?)`, nullableInt(userID), hashToken(token), kind, string(data), expires.Unix())
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return false
	}
	http.SetCookie(w, webAuthnChallengeCookie(token, int(time.Until(expires).Seconds())))
	return true
}

// Returns the WebAuthn SessionData for the current request
func loadWebAuthnChallenge(r *http.Request, db *sql.DB, userID *int, kind string) (webauthn.SessionData, int, error) {
	var waSession webauthn.SessionData

	// Get the cookie and load the session from database
	cookie, err := r.Cookie("webauthn_challenge")
	if err != nil || cookie.Value == "" {
		return waSession, 0, sql.ErrNoRows
	}

	// Query the latest non-expired session for this user and kind, if userID is provided. Otherwise, query the latest session by token and kind.
	challenge, err := sqlScanOne(r, db, `
		SELECT id, session_data
		FROM webauthn_challenges
		WHERE (? IS NULL OR user_id = ?)
			AND token_hash = ?
			AND kind = ?
			AND expires_at > strftime('%s','now')
		ORDER BY id DESC
		LIMIT 1`, model.ScanWebAuthnChallenge, nullableInt(userID), nullableInt(userID), hashToken(cookie.Value), kind)
	if err != nil {
		return waSession, 0, err
	}

	// Deserialize the session data from JSON
	err = json.Unmarshal([]byte(challenge.SessionData), &waSession)
	return waSession, challenge.ID, err
}

// webAuthnChallengeCookie builds the short-lived challenge cookie.
func webAuthnChallengeCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "webauthn_challenge",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}

// nullableInt converts an optional int to a SQL NULL-compatible value.
func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
