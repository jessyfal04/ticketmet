package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func runEvery(ctx context.Context, interval time.Duration, run func()) {
	run()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func runChan[T any](ctx context.Context, c <-chan T, handle func(T)) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-c:
			if !ok {
				return
			}
			handle(item)
		}
	}
}

// HTTP GET with query params and header , then decodes JSON response
func getJSONQuery(ctx context.Context, client *http.Client, rawURL string, query url.Values, headers map[string]string, payload any, allowNotFound bool) error {
	// Append query params when we have some
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}

	// Build the request with the caller context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	// Set headers for the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request and close the body
	log.Printf("[http] GET %s", rawURL)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[http] GET failed %s: %v", rawURL, err)
		return err
	}
	defer resp.Body.Close()

	// If 404 is allowed, return nil for not found
	if resp.StatusCode == http.StatusNotFound && allowNotFound {
		return nil
	}

	// Any non-2xx status is an error.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http status %d", resp.StatusCode)
	}

	// Decode the JSON payload into payload
	return json.NewDecoder(resp.Body).Decode(payload)
}

// Helpers for env vars
func Getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func GetenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
