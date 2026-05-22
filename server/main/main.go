package main

import (
	"context"
	"database/sql"
	_ "embed"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"server/api"
	"server/job"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string // Put the content of schema.sql in the string

func main() {
	// Log file
	logFile, err := configureLog(getenv("LOG_PATH", "data/ticketmet.log"))
	if err != nil {
		log.Fatalf("Log error: %v", err)
	}
	defer logFile.Close()

	// Database
	db, err := openDB(getenv("DB_PATH", "data/ticketmet.sqlite3"))
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	// Start Ticketmaster sync
	go job.RunTicketmaster(context.Background(), db, getenv("TICKETMASTER_API_KEY", ""), 15*time.Minute)

	// Start mailserver
	mailServer := job.NewFromEnv()
	go mailServer.Run(context.Background())

	// Start alert radar
	go job.RunAlertRadar(context.Background(), db, mailServer.C, time.Minute)

	// Start Setlist.fm
	setlistServer := job.NewSetlistFMServer(db, getenv("SETLISTFM_API_KEY", ""))
	go setlistServer.Run(context.Background())

	// Start HTTP server
	clientDir := getenv("CLIENT_DIR", "../client")
	mux := api.ServeMux(db, clientDir, mailServer.C, setlistServer.C)

	addr := ":" + getenv("PORT", "8080")
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func configureLog(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	_ = os.Remove(path)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	return file, nil
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	// Remove existing database to start fresh.
	// _ = os.Remove(path)
	// _ = os.Remove(path + "-shm")
	// _ = os.Remove(path + "-wal")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
