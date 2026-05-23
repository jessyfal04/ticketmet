package job

import (
	"context"
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
