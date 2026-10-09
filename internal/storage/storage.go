package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	connectAttempts = 10
	connectDelay    = 2 * time.Second
)

// Open creates the shared connection pool and waits for the database to
// answer, so a Postgres container that is still starting does not crash-loop
// the app. Call it once at startup and close the pool on shutdown.
func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	var pingErr error
	for attempt := 1; attempt <= connectAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr = db.PingContext(pingCtx)
		cancel()
		if pingErr == nil {
			return db, nil
		}

		select {
		case <-ctx.Done():
			db.Close()
			return nil, ctx.Err()
		case <-time.After(connectDelay):
		}
	}
	db.Close()

	return nil, fmt.Errorf("database not reachable after %d attempts: %w", connectAttempts, pingErr)
}
