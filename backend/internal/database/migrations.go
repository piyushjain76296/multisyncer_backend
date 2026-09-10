package database

import (
	"context"
	"database/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS windows (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS media (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			url TEXT NOT NULL,
			duration_seconds INT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS playlist_items (
			id UUID PRIMARY KEY,
			window_id UUID NOT NULL REFERENCES windows(id) ON DELETE CASCADE,
			media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			position INT NOT NULL,
			duration_seconds INT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_playlist_items_window_pos ON playlist_items(window_id, position);`,
		`CREATE TABLE IF NOT EXISTS sync_sessions (
			id UUID PRIMARY KEY,
			media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			start_at TIMESTAMP WITH TIME ZONE NOT NULL,
			end_at TIMESTAMP WITH TIME ZONE NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sync_sessions_dates ON sync_sessions(start_at, end_at);`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
