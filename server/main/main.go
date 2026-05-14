package main

import (
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"server/api"
	"server/job"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string // Put the content of schema.sql in the string

func main() {
	db, err := openDB(getenv("DB_PATH", "data/ticketmet.sqlite3"))
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	clientDir := getenv("CLIENT_DIR", "../client")
	mux := api.ServeMux(db, clientDir)

	job.StartTicketmaster(db, os.Getenv("TICKETMASTER_API_KEY"))

	addr := ":" + getenv("PORT", "8080")
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	// Remove existing database to start fresh.
	_ = os.Remove(path)
	_ = os.Remove(path + "-shm")
	_ = os.Remove(path + "-wal")

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
