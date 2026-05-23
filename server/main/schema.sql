PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS artists (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS venues (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	city TEXT NOT NULL,
	country TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS concerts (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	date TEXT NOT NULL,
	venue_id INTEGER NOT NULL REFERENCES venues(id),
	artist_id INTEGER NOT NULL REFERENCES artists(id),
	url TEXT NOT NULL,
	photo_url TEXT NOT NULL DEFAULT '',
	seatmap_url TEXT NOT NULL DEFAULT '',
	sale_start_datetime TEXT NOT NULL,
	created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL UNIQUE,
	expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS webauthn_credentials (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	credential_id TEXT NOT NULL UNIQUE,
	public_key TEXT NOT NULL,
	sign_count INTEGER NOT NULL DEFAULT 0,
	user_present BOOLEAN NOT NULL DEFAULT FALSE,
	user_verified BOOLEAN NOT NULL DEFAULT FALSE,
	backup_eligible BOOLEAN NOT NULL DEFAULT FALSE,
	backup_state BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS webauthn_challenges (
	id INTEGER PRIMARY KEY,
	user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL UNIQUE,
	kind TEXT NOT NULL CHECK (kind IN ('registration', 'login')),
	session_data TEXT NOT NULL,
	expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS user_sns (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	sns TEXT NOT NULL,
	PRIMARY KEY (user_id, sns)
);

CREATE TABLE IF NOT EXISTS wt (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	concert_id INTEGER NOT NULL REFERENCES concerts(id) ON DELETE CASCADE,
	type TEXT NOT NULL CHECK (type IN ('wtb', 'wts')),
	PRIMARY KEY (user_id, concert_id)
);

CREATE TABLE IF NOT EXISTS favorites (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	concert_id INTEGER NOT NULL REFERENCES concerts(id) ON DELETE CASCADE,
	created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
	PRIMARY KEY (user_id, concert_id)
);

CREATE TABLE IF NOT EXISTS alerts (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	target_type TEXT NOT NULL CHECK (target_type IN ('artist', 'venue')),
	target_id INTEGER NOT NULL,
	created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
	UNIQUE(user_id, target_type, target_id)
);

CREATE TABLE IF NOT EXISTS notifications (
	dedupe_key TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS setlists (
	concert_id INTEGER NOT NULL REFERENCES concerts(id) ON DELETE CASCADE,
	song_order INTEGER NOT NULL,
	song_name TEXT NOT NULL,
	PRIMARY KEY (concert_id, song_order)
);

CREATE TABLE IF NOT EXISTS sync_ticketmaster (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	max_visibility TEXT NOT NULL
);

INSERT OR IGNORE INTO sync_ticketmaster (id, max_visibility)
VALUES (1, '');
