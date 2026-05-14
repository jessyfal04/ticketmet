package api

import (
	"database/sql"
	"net/http"
	"server/model"
	"strings"
)

type profileResponse struct {
	User      model.PublicUser
	SNS       []string
	Favorites []model.DisplayConcert
	WT        []model.ProfileWT
	Alerts    []alertResponse
}

// Return the full current-user application profile.
func handleProfileGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	profile, err := loadProfile(r, db, user)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	writeJSON(w, profile)
}

// Update the current-user SNS handles.
func handleProfilePatch(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user, ok := requireUser(w, r, db)
	if !ok {
		return
	}

	var body struct {
		SNS []string
	}
	if !readJSON(w, r, &body) {
		return
	}

	if err := sqlExec(r, db, "DELETE FROM user_sns WHERE user_id = ?", user.ID); err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}

	seen := map[string]bool{}
	for _, raw := range body.SNS {
		sns := strings.TrimSpace(raw)
		key := strings.ToLower(sns)
		if sns == "" || seen[key] {
			continue
		}
		seen[key] = true
		if err := sqlExec(r, db, "INSERT INTO user_sns (user_id, sns) VALUES (?, ?)", user.ID, sns); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
	}

	profile, err := loadProfile(r, db, user)
	if err != nil {
		logHttpError(w, http.StatusInternalServerError, "", err)
		return
	}
	writeJSON(w, profile)
}

func loadProfile(r *http.Request, db *sql.DB, user model.User) (profileResponse, error) {
	// Query SNS handles
	sns, err := sqlScanList(r, db, `
		SELECT sns
		FROM user_sns
		WHERE user_id = ?
		ORDER BY sns`, model.ScanString, user.ID)
	if err != nil {
		return profileResponse{}, err
	}

	// Query favorite concerts
	favorites, err := sqlScanList(r, db, `
		SELECT c.id,
			c.name,
			c.date,
			c.venue_id,
			c.artist_id,
			c.url,
			c.photo_url,
			c.seatmap_url,
			c.sale_start_datetime,
			v.name,
			a.name
		FROM concerts c
		JOIN venues v ON v.id = c.venue_id
		JOIN artists a ON a.id = c.artist_id
		JOIN favorites f ON f.concert_id = c.id
		WHERE f.user_id = ?
		ORDER BY c.date`, model.ScanDisplayConcert, user.ID)
	if err != nil {
		return profileResponse{}, err
	}

	// Query WTB/WTS concerts
	wtItems, err := sqlScanList(r, db, `
		SELECT wt.type,
			c.id,
			c.name,
			c.date,
			c.venue_id,
			c.artist_id,
			c.url,
			c.photo_url,
			c.seatmap_url,
			c.sale_start_datetime,
			v.name,
			a.name
		FROM wt
		JOIN concerts c ON c.id = wt.concert_id
		JOIN venues v ON v.id = c.venue_id
		JOIN artists a ON a.id = c.artist_id
		WHERE wt.user_id = ?
		ORDER BY c.date`, model.ScanProfileWT, user.ID)
	if err != nil {
		return profileResponse{}, err
	}

	// Query alerts with their display name
	alerts, err := sqlScanList(r, db, `
		SELECT al.id,
			al.target_type,
			al.target_id,
			COALESCE(ar.name, ve.name, '')
		FROM alerts al
		LEFT JOIN artists ar ON al.target_type = 'artist' AND ar.id = al.target_id
		LEFT JOIN venues ve ON al.target_type = 'venue' AND ve.id = al.target_id
		WHERE al.user_id = ?
		ORDER BY al.id`, func(row interface{ Scan(...any) error }) (alertResponse, error) {
		var alert alertResponse
		err := row.Scan(&alert.ID, &alert.TargetType, &alert.TargetID, &alert.TargetName)
		return alert, err
	}, user.ID)
	if err != nil {
		return profileResponse{}, err
	}

	return profileResponse{
		User:      user.Public(),
		SNS:       sns,
		Favorites: favorites,
		WT:        wtItems,
		Alerts:    alerts,
	}, nil
}


