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
	sale_start_datetime TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	username TEXT NOT NULL UNIQUE
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
	PRIMARY KEY (user_id, concert_id, type)
);

CREATE TABLE IF NOT EXISTS favorites (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	concert_id INTEGER NOT NULL REFERENCES concerts(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, concert_id)
);

CREATE TABLE IF NOT EXISTS alerts (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	target_type TEXT NOT NULL,
	target_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS setlists (
	concert_id INTEGER NOT NULL REFERENCES concerts(id) ON DELETE CASCADE,
	song_order INTEGER NOT NULL,
	song_name TEXT NOT NULL,
	PRIMARY KEY (concert_id, song_order)
);

CREATE TABLE IF NOT EXISTS sync_ticketmaster (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	last_public_visibility_start_datetime TEXT NOT NULL
);
