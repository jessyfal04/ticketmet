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

const (
	ticketmasterInterval = 15 * time.Minute
	alertRadarInterval   = time.Minute
)

func main() {
	// Log file
	logFile, err := configureLog(job.Getenv("LOG_PATH", "data/ticketmet.log"))
	if err != nil {
		log.Fatalf("Log error: %v", err)
	}
	defer logFile.Close()

	// Database
	db, err := openDB(job.Getenv("DB_PATH", "data/ticketmet.sqlite3"))
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	// Start DB server
	dbChan := make(chan job.DBRequest, 64)
	job.RunDBServer(context.Background(), db, dbChan)

	// Start Ticketmaster sync
	go job.RunTicketmaster(context.Background(), dbChan, job.Getenv("TICKETMASTER_API_KEY", ""), ticketmasterInterval)

	// Start mailserver
	mailChan := make(chan job.Envelope, 64)
	job.RunMailServer(context.Background(), job.Config{
		Host: job.Getenv("SMTP_HOST", "10.66.66.1"),
		Port: job.GetenvInt("SMTP_PORT", 25),
		From: job.Getenv("SMTP_FROM", "ticketmet@jessyfal04.dev"),
	}, mailChan)

	// Start alert radar
	go job.RunAlertRadar(context.Background(), dbChan, mailChan, alertRadarInterval)

	// Start Setlist.fm
	setlistChan := make(chan job.SetlistRequest, 32)
	job.RunSetlistFMServer(context.Background(), dbChan, job.Getenv("SETLISTFM_API_KEY", ""), setlistChan)

	// Start HTTP server
	clientDir := job.Getenv("CLIENT_DIR", "../client")
	mux := api.ServeMux(clientDir, dbChan, mailChan, setlistChan)

	addr := ":" + job.Getenv("PORT", "8080")
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Configure logging to write to both stdout and a file, with timestamps
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

// Open the SQLite database, creating it if it doesn't exist, and apply the schema
func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	// Remove existing database to start fresh.
	if os.Getenv("ERASE_DB") == "1" {
		_ = os.Remove(path)
		_ = os.Remove(path + "-shm")
		_ = os.Remove(path + "-wal")
	}

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
