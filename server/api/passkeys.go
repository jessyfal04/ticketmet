package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"server/job"
	"server/model"
	"strconv"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Model user.Model x webauthn.User
type webAuthnUser struct {
	model.User
	credentials []webauthn.Credential
}

type passkeyRegisterResponse struct {
	OK bool
}

type passkeyLoginResponse struct {
	User model.PublicUser
}

type passkeyListResponse struct {
	Passkeys []model.PublicPasskey
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
func handlePasskeyRegisterBegin(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load current user
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		// Build registration options
		wuser, err := loadWebAuthnUser(r, dbChan, user.ID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		wa, err := newWebAuthn(r)
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
		if !saveWebAuthnChallenge(w, r, dbChan, &user.ID, "registration", wa_sessionData) {
			return
		}
		writeJSON(w, options)
	}
}

// Store the created passkey after browser verification
func handlePasskeyRegisterFinish(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load current user
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		// Load the WebAuthn user
		wuser, err := loadWebAuthnUser(r, dbChan, user.ID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		// Load the WebAuthn wa_sessionData saved during begin
		wa_sessionData, sessionID, err := loadWebAuthnChallenge(r, dbChan, &user.ID, "registration")
		if err != nil {
			logHttpError(w, http.StatusBadRequest, "passkey registration challenge not found", err)
			return
		}
		wa, err := newWebAuthn(r)
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
		err = job.SqlExec(r.Context(), dbChan, `
		INSERT INTO webauthn_credentials (user_id, credential_id, public_key, sign_count, user_present,	user_verified, backup_eligible, backup_state
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			user.ID,
			base64.RawURLEncoding.EncodeToString(credential.ID),
			base64.RawURLEncoding.EncodeToString(credential.PublicKey),
			credential.Authenticator.SignCount,
			credential.Flags.UserPresent,
			credential.Flags.UserVerified,
			credential.Flags.BackupEligible,
			credential.Flags.BackupState)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		// Delete the WebAuthn session
		err = job.SqlExec(r.Context(), dbChan, "DELETE FROM webauthn_challenges WHERE id = ?", sessionID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		http.SetCookie(w, webAuthnChallengeCookie(r, "", -1))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, passkeyRegisterResponse{OK: true})
	}
}

// Start passkey login
func handlePasskeyLoginBegin(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build login options, no need email
		wa, err := newWebAuthn(r)
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
		if !saveWebAuthnChallenge(w, r, dbChan, nil, "login", wa_sessionData) {
			return
		}
		writeJSON(w, options)
	}
}

// Finish the login with the passkey
func handlePasskeyLoginFinish(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load the WebAuthn wa_sessionData saved during begin
		wa_sessionData, sessionID, err := loadWebAuthnChallenge(r, dbChan, nil, "login")
		if err != nil {
			logHttpError(w, http.StatusBadRequest, "passkey login challenge not found", err)
			return
		}

		// Build WebAuthn config
		wa, err := newWebAuthn(r)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		var user webAuthnUser
		var credential *webauthn.Credential

		// Verify credential
		resolvedUser, credential, err := wa.FinishPasskeyLogin(
			func(rawID, userHandle []byte) (webauthn.User, error) {
				if len(rawID) > 0 {
					return loadWebAuthnUserByCredentialID(r, dbChan, rawID)
				}
				id, err := strconv.Atoi(string(userHandle))
				if err != nil {
					return nil, err
				}
				return loadWebAuthnUser(r, dbChan, id)
			}, wa_sessionData, r)
		if err != nil {
			logHttpError(w, http.StatusBadRequest, "invalid passkey", err)
			return
		}
		// Type assert to our webAuthnUser model
		resolved, ok := resolvedUser.(webAuthnUser)
		if !ok {
			logHttpError(w, http.StatusBadRequest, "invalid passkey user", nil)
			return
		}
		user = resolved

		// Update mutable credential state and remove the one-time WebAuthn wa_sessionData.
		err = job.SqlExec(r.Context(), dbChan, `
		UPDATE webauthn_credentials
		SET sign_count = ?,
			user_present = ?,
			user_verified = ?,
			backup_state = ?
		WHERE credential_id = ?`,
			credential.Authenticator.SignCount,
			credential.Flags.UserPresent,
			credential.Flags.UserVerified,
			credential.Flags.BackupState,
			base64.RawURLEncoding.EncodeToString(credential.ID))
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		// Delete the WebAuthn session
		err = job.SqlExec(r.Context(), dbChan, "DELETE FROM webauthn_challenges WHERE id = ?", sessionID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		http.SetCookie(w, webAuthnChallengeCookie(r, "", -1))
		if !createSessionCookie(w, r, dbChan, user.ID) {
			return
		}
		writeJSON(w, passkeyLoginResponse{User: user.Public()})
	}
}

// List current-user passkeys
func handlePasskeyList(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load current user
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		// Query credentials
		items, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT credential_id,
			public_key,
			sign_count,
			user_present,
			user_verified,
			backup_eligible,
			backup_state
		FROM webauthn_credentials
		WHERE user_id = ?
		ORDER BY id`, model.ScanPublicPasskey, user.ID)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		writeJSON(w, passkeyListResponse{Passkeys: items})
	}
}

// Delete one current-user passkey
func handlePasskeyDelete(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load current user
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		// Delete credential
		credentialId, ok := httpGetStringParam(w, r, "credentialId")
		if !ok {
			return
		}

		err := job.SqlExec(r.Context(), dbChan, "DELETE FROM webauthn_credentials WHERE user_id = ? AND credential_id = ?", user.ID, credentialId)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Helper

// newWebAuthn builds WebAuthn config for prod and local dev.
func newWebAuthn(r *http.Request) (*webauthn.WebAuthn, error) {
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	rpID := strings.Split(host, ":")[0]
	if value := os.Getenv("WEBAUTHN_RP_ID"); value != "" {
		rpID = value
	}

	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		proto = forwardedProto
	}
	origin := proto + "://" + host

	origins := []string{origin}
	if value := os.Getenv("WEBAUTHN_ORIGINS"); value != "" {
		origins = strings.Split(value, ",")
	}

	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: "TicketMet",
		RPOrigins:     origins,
	})
}

// Load a user with its credentials
func loadWebAuthnUser(r *http.Request, dbChan chan<- job.DBRequest, id int) (webAuthnUser, error) {
	user, err := job.SqlScanOne(r.Context(), dbChan, `
		SELECT id, email, password_hash
		FROM users
		WHERE id = ?`, model.ScanUser, id)
	if err != nil {
		return webAuthnUser{}, err
	}

	passkeys, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT credential_id,
			public_key,
			sign_count,
			user_present,
			user_verified,
			backup_eligible,
			backup_state
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
			ID:        id,
			PublicKey: publicKey,
			Flags: webauthn.CredentialFlags{
				UserPresent:    passkey.UserPresent,
				UserVerified:   passkey.UserVerified,
				BackupEligible: passkey.BackupEligible,
				BackupState:    passkey.BackupState,
			},
			Authenticator: webauthn.Authenticator{SignCount: uint32(passkey.SignCount)},
		})
	}
	return webAuthnUser{
		User:        user,
		credentials: credentials,
	}, nil
}

// Load the owner of a credential ID, then load all their credentials.
func loadWebAuthnUserByCredentialID(r *http.Request, dbChan chan<- job.DBRequest, credentialID []byte) (webAuthnUser, error) {
	encodedID := base64.RawURLEncoding.EncodeToString(credentialID)

	user, err := job.SqlScanOne(r.Context(), dbChan, `
		SELECT u.id, u.email, u.password_hash
		FROM users u
		JOIN webauthn_credentials wc ON wc.user_id = u.id
		WHERE wc.credential_id = ?`, model.ScanUser, encodedID)
	if err != nil {
		return webAuthnUser{}, err
	}

	return loadWebAuthnUser(r, dbChan, user.ID)
}

// Stores the WebAuthn SessionData
func saveWebAuthnChallenge(w http.ResponseWriter, r *http.Request, dbChan chan<- job.DBRequest, userID *int, kind string, wa_sessionData *webauthn.SessionData) bool {
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
	err = job.SqlExec(r.Context(), dbChan, `
		INSERT INTO webauthn_challenges (user_id, token_hash, kind, session_data, expires_at)
		VALUES (?, ?, ?, ?, ?)`, nullableInt(userID), hashToken(token), kind, string(data), expires.Unix())
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return false
	}
	http.SetCookie(w, webAuthnChallengeCookie(r, token, int(time.Until(expires).Seconds())))
	return true
}

// Returns the WebAuthn SessionData for the current request
func loadWebAuthnChallenge(r *http.Request, dbChan chan<- job.DBRequest, userID *int, kind string) (webauthn.SessionData, int, error) {
	var waSession webauthn.SessionData

	// Get the cookie and load the session from database
	cookie, err := r.Cookie("webauthn_challenge")
	if err != nil || cookie.Value == "" {
		return waSession, 0, sql.ErrNoRows
	}

	// Query the latest non-expired session for this user and kind, if userID is provided. Otherwise, query the latest session by token and kind.
	challenge, err := job.SqlScanOne(r.Context(), dbChan, `
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
func webAuthnChallengeCookie(r *http.Request, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "webauthn_challenge",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
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
