package job

import (
	"context"
	"fmt"
	"log"
	"server/model"
	"strconv"
	"strings"
	"time"
)

// Stats
// AlertRadarStats summarizes one alert radar pass
type AlertRadarStats struct {
	StartedAt       time.Time
	AlertCandidates int
	SaleCandidates  int
	TradeCandidates int
	Users           int
	QueueFull       int
}

// Stats string output for logs
func (s AlertRadarStats) String() string {
	return fmt.Sprintf(
		"[alertRadar] stats: alert_candidates=%d sale_candidates=%d trade_candidates=%d users=%d queue_full=%d duration=%s",
		s.AlertCandidates,
		s.SaleCandidates,
		s.TradeCandidates,
		s.Users,
		s.QueueFull,
		time.Since(s.StartedAt).Round(time.Millisecond),
	)
}

// RunAlertRadar periodically detects alert and WTB/WTS matches and queues mails.
func RunAlertRadar(ctx context.Context, dbChan chan<- DBRequest, mailChan chan<- Envelope, interval time.Duration) {
	runEvery(ctx, interval, func() {
		stats, err := AlertRadarPass(ctx, dbChan, mailChan)
		if err != nil {
			log.Printf("[alertRadar] failed: %v", err)
			return
		}
		log.Print(stats)
	})
}

type mailGroupInfo struct {
	Email          string
	DedupeKeys     []string
	AlertMailItems []AlertMailItem
}

// One pass of radar
func AlertRadarPass(ctx context.Context, dbChan chan<- DBRequest, mailChan chan<- Envelope) (AlertRadarStats, error) {
	// Stats with start time
	stats := AlertRadarStats{StartedAt: time.Now().UTC()}

	// We get the concert alert : users with an alert
	alertCandidates, err := loadConcertAlertCandidates(ctx, dbChan)
	if err != nil {
		return stats, err
	}
	stats.AlertCandidates = len(alertCandidates)

	// We get the sale alert : favorites with open sale
	saleCandidates, err := loadFavoriteAlertCandidates(ctx, dbChan)
	if err != nil {
		return stats, err
	}
	stats.SaleCandidates = len(saleCandidates)

	// We get the trade alert : WTB/WTS matches
	tradeCandidates, err := loadTradeAlertCandidates(ctx, dbChan)
	if err != nil {
		return stats, err
	}
	stats.TradeCandidates = len(tradeCandidates)

	// We groupe the alert by userID
	mailGroups := map[int]*mailGroupInfo{}
	candidates := append(alertCandidates, saleCandidates...)
	candidates = append(candidates, tradeCandidates...)
	for _, candidate := range candidates {
		// We create the group if not exists
		mailGroup := mailGroups[candidate.UserID]
		if mailGroup == nil {
			mailGroup = &mailGroupInfo{Email: candidate.Email}
			mailGroups[candidate.UserID] = mailGroup
		}

		// We add the dedupekey+item in the group
		mailGroup.DedupeKeys = append(mailGroup.DedupeKeys, candidate.DedupeKey)
		mailGroup.AlertMailItems = append(mailGroup.AlertMailItems, AlertMailItem{
			Title:   candidate.Title,
			Details: candidate.Details,
			URL:     candidate.URL,
		})
	}

	// We send one mail per user with all their alerts
	for _, mailGroup := range mailGroups {
		envelope := Envelope{
			Dst:     mailGroup.Email,
			Message: AlertMail(mailGroup.Email, mailGroup.AlertMailItems),
		}

		// We try queueing the mail
		select {
		// ok
		case mailChan <- envelope:
			// We insert the dedupe keys
			for _, key := range mailGroup.DedupeKeys {
				if err := SqlExec(ctx, dbChan, `INSERT OR IGNORE INTO notifications (dedupe_key) VALUES (?)`, key); err != nil {
					return stats, err
				}
			}
			stats.Users++
			log.Printf("[alertRadar] mail queued for %s items=%d", mailGroup.Email, len(mailGroup.AlertMailItems))

		// full
		default:
			stats.QueueFull++
			log.Printf("[alertRadar] mail queue full for %s items=%d", mailGroup.Email, len(mailGroup.AlertMailItems))
		}
	}

	return stats, nil
}

// Potential alert
type alertCandidate struct {
	UserID    int
	Email     string
	DedupeKey string
	Title     string
	Details   string
	URL       string
}

// Wrapable function to load alert candidates with any query and scan func
func loadAlertCandidates[T any](ctx context.Context, dbChan chan<- DBRequest, query string, scan ScanFunc[T], build func(T) alertCandidate, args ...any) ([]alertCandidate, error) {
	// Execute the query and scan the results
	rows, err := SqlScanList(ctx, dbChan, query, scan, args...)
	if err != nil {
		return nil, err
	}

	// Convert the results by building alert candidates
	candidates := make([]alertCandidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, build(row))
	}
	return candidates, nil
}

// Check if there is a new concert matching the user's alerts (artist/venue)
func loadConcertAlertCandidates(ctx context.Context, dbChan chan<- DBRequest) ([]alertCandidate, error) {
	return loadAlertCandidates(ctx, dbChan, `
		WITH candidates AS (
			SELECT u.id AS user_id, u.email AS email, 'concert_alert:' || al.target_type || ':user=' || u.id || ':target=' || al.target_id || ':concert=' || c.id AS dedupe_key, c.id AS concert_id, c.name AS concert_name, a.name AS artist_name, v.name AS venue_name, v.city AS venue_city, v.country AS venue_country, c.date AS concert_date, al.target_type AS target_type
			FROM alerts al
			JOIN users u ON u.id = al.user_id
			JOIN concerts c ON (al.target_type = 'artist' AND c.artist_id = al.target_id) OR (al.target_type = 'venue' AND c.venue_id = al.target_id)
			JOIN artists a ON a.id = c.artist_id
			JOIN venues v ON v.id = c.venue_id
			WHERE c.created_at >= al.created_at
		)
		SELECT ca.user_id, ca.email, ca.dedupe_key, ca.concert_id, ca.concert_name, ca.artist_name, ca.venue_name, ca.concert_date, ca.target_type
		FROM candidates ca
		LEFT JOIN notifications n ON n.dedupe_key = ca.dedupe_key
		WHERE n.dedupe_key IS NULL
		ORDER BY ca.user_id, ca.concert_date`,
		model.ScanAlertConcertCandidate, func(row model.AlertConcertCandidate) alertCandidate {
			// Detection of artist/venue to adapt the message
			title := fmt.Sprintf("New concert by %s", row.ArtistName)
			if row.TargetType == "venue" {
				title = fmt.Sprintf("New concert at %s", row.VenueName)
			}

			return alertCandidate{
				UserID:    row.UserID,
				Email:     row.Email,
				DedupeKey: row.DedupeKey,
				Title:     title,
				Details:   readableConcert(row.ConcertName, row.ArtistName, row.VenueName, row.ConcertDate),
				URL:       concertLink(row.ConcertID),
			}
		})
}

// Check if there is a new open sale for the user's favorite concerts
func loadFavoriteAlertCandidates(ctx context.Context, dbChan chan<- DBRequest) ([]alertCandidate, error) {
	return loadAlertCandidates(ctx, dbChan, `
		WITH candidates AS (
			SELECT u.id AS user_id, u.email AS email, 'sale_alert:user=' || u.id || ':concert=' || c.id AS dedupe_key, c.id AS concert_id, c.name AS concert_name, a.name AS artist_name, v.name AS venue_name, c.date AS concert_date, c.sale_start_datetime AS sale_start_datetime
			FROM favorites f
			JOIN users u ON u.id = f.user_id
			JOIN concerts c ON c.id = f.concert_id
			JOIN artists a ON a.id = c.artist_id
			JOIN venues v ON v.id = c.venue_id
			WHERE c.sale_start_datetime <> '' AND f.created_at < CAST(strftime('%s', c.sale_start_datetime) AS INTEGER) AND c.sale_start_datetime <= ?
		)
		SELECT ca.user_id, ca.email, ca.dedupe_key, ca.concert_id, ca.concert_name, ca.artist_name, ca.venue_name, ca.concert_date
		FROM candidates ca
		LEFT JOIN notifications n ON n.dedupe_key = ca.dedupe_key
		WHERE n.dedupe_key IS NULL
		ORDER BY ca.user_id, ca.sale_start_datetime`,
		model.ScanAlertFavoriteCandidate, func(row model.AlertFavoriteCandidate) alertCandidate {
			return alertCandidate{
				UserID:    row.UserID,
				Email:     row.Email,
				DedupeKey: row.DedupeKey,
				Title:     fmt.Sprintf("Sale open: %s", row.ConcertName),
				Details:   "Sale is open.\n" + readableConcert(row.ConcertName, row.ArtistName, row.VenueName, row.ConcertDate),
				URL:       concertLink(row.ConcertID),
			}
		}, time.Now().UTC().Format(time.RFC3339))
}

// Check if there is a new WTB/WTS match for the user's open WTB/WTS
func loadTradeAlertCandidates(ctx context.Context, dbChan chan<- DBRequest) ([]alertCandidate, error) {
	return loadAlertCandidates(ctx, dbChan, `
		WITH candidates AS (
			SELECT u.id AS user_id, u.email AS email, 'trade_match:user=' || u.id || ':concert=' || c.id || ':self=' || wt.type || ':peer=' || peer.user_id AS dedupe_key, c.id AS concert_id, c.name AS concert_name, a.name AS artist_name, v.name AS venue_name, c.date AS concert_date, wt.type AS self_type, peer.type AS peer_type
			FROM wt
			JOIN wt peer ON peer.concert_id = wt.concert_id AND peer.type <> wt.type AND peer.user_id <> wt.user_id
			JOIN users u ON u.id = wt.user_id
			JOIN concerts c ON c.id = wt.concert_id
			JOIN artists a ON a.id = c.artist_id
			JOIN venues v ON v.id = c.venue_id
		)
		SELECT ca.user_id, ca.email, ca.dedupe_key, ca.concert_id, ca.concert_name, ca.artist_name, ca.venue_name, ca.concert_date, ca.self_type, ca.peer_type
		FROM candidates ca
		LEFT JOIN notifications n ON n.dedupe_key = ca.dedupe_key
		WHERE n.dedupe_key IS NULL
		ORDER BY ca.user_id, ca.concert_date`,
		model.ScanAlertTradeCandidate, func(row model.AlertTradeCandidate) alertCandidate {

			// Detection of WTB/WTS to adapt the message
			details := "Someone is selling a ticket for a concert you want."
			if row.SelfType == "wts" && row.PeerType == "wtb" {
				details = "Someone is looking for a ticket for a concert you are selling."
			}

			return alertCandidate{
				UserID:    row.UserID,
				Email:     row.Email,
				DedupeKey: row.DedupeKey,
				Title:     fmt.Sprintf("Match %s/%s: %s", row.SelfType, row.PeerType, row.ConcertName),
				Details:   details + "\n" + readableConcert(row.ConcertName, row.ArtistName, row.VenueName, row.ConcertDate),
				URL:       concertLink(row.ConcertID),
			}
		})
}

// Helpeurs

// Give the link to the concert page
func concertLink(concertID int) string {
	return strings.TrimRight(Getenv("APP_BASE_URL", "https://ticketmet.jessyfal04.dev"), "/") + "/?concert=" + strconv.Itoa(concertID)
}

// Convert concert details to a human readable format.
func readableConcert(concertName string, artistName string, venueName string, concertDate string) string {
	parts := []string{concertName, artistName, venueName, readableTime(concertDate)}
	return strings.Join(parts, " • ")
}

// Time to human readable
func readableTime(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return parsed.Format("02/01/2006 15:04")
}
