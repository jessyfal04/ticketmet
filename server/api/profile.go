package api

import (
	"net/http"
	"server/job"
	"server/model"
	"strings"
)

type profileResponse struct {
	User      model.PublicUser
	SNS       []string
	Favorites []model.DisplayConcert
	WT        []model.ProfileWT
	Alerts    []model.ProfileAlert
}

// Return the full current-user application profile.
func handleProfileGet(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		profile, err := loadProfile(r, dbChan, user)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, profile)
	}
}

// Update the current-user SNS
func handleProfilePatch(dbChan chan<- job.DBRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireUser(w, r, dbChan)
		if !ok {
			return
		}

		var body struct {
			SNS []string
		}
		if !readJSON(w, r, &body) {
			return
		}

		if err := job.SqlExec(r.Context(), dbChan, "DELETE FROM user_sns WHERE user_id = ?", user.ID); err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}

		// SNS decoding
		seen := map[string]bool{}
		for _, raw := range body.SNS {
			sns := strings.TrimSpace(raw)
			key := strings.ToLower(sns)
			if sns == "" || seen[key] {
				continue
			}
			seen[key] = true
			if err := job.SqlExec(r.Context(), dbChan, "INSERT INTO user_sns (user_id, sns) VALUES (?, ?)", user.ID, sns); err != nil {
				logHttpError(w, http.StatusInternalServerError, "", err)
				return
			}
		}

		profile, err := loadProfile(r, dbChan, user)
		if err != nil {
			logHttpError(w, http.StatusInternalServerError, "", err)
			return
		}
		writeJSON(w, profile)
	}
}

func loadProfile(r *http.Request, dbChan chan<- job.DBRequest, user model.User) (profileResponse, error) {
	// Query SNS handles
	sns, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT sns
		FROM user_sns
		WHERE user_id = ?
		ORDER BY sns`, model.ScanString, user.ID)
	if err != nil {
		return profileResponse{}, err
	}

	// Query favorite concerts
	favorites, err := job.SqlScanList(r.Context(), dbChan, `
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
	wtItems, err := job.SqlScanList(r.Context(), dbChan, `
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
	alerts, err := job.SqlScanList(r.Context(), dbChan, `
		SELECT al.id,
			al.target_type,
			al.target_id,
			COALESCE(ar.name, ve.name, '')
		FROM alerts al
		LEFT JOIN artists ar ON al.target_type = 'artist' AND ar.id = al.target_id
		LEFT JOIN venues ve ON al.target_type = 'venue' AND ve.id = al.target_id
		WHERE al.user_id = ?
		ORDER BY al.id`, model.ScanProfileAlert, user.ID)
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
