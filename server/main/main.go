package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"server/api"
	"server/model"
)

func main() {
	data := model.DummyData()
	clientDir := getenv("CLIENT_DIR", "../client")
	mux := api.ServeMux(data, clientDir)

	addr := ":" + getenv("PORT", "8080")
	fmt.Printf("Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Return the value of the environment variable `key` if it exists, otherwise return `fallback`.
func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
