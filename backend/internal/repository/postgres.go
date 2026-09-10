package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/user/backend_multisyncer/internal/models"
)

var ErrNotFound = errors.New("resource not found")

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// --- Windows ---

func (r *Repository) GetWindows(ctx context.Context) ([]models.Window, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, created_at, updated_at FROM windows ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []models.Window
	for rows.Next() {
		var w models.Window
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	return windows, nil
}

func (r *Repository) CreateWindow(ctx context.Context, name string) (*models.Window, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := r.DB.ExecContext(ctx, "INSERT INTO windows (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		id, name, now, now)
	if err != nil {
		return nil, err
	}
	return &models.Window{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *Repository) DeleteWindow(ctx context.Context, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, "DELETE FROM windows WHERE id = $1", id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Media ---

func (r *Repository) GetMedia(ctx context.Context) ([]models.Media, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, type, url, duration_seconds, created_at, updated_at FROM media ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []models.Media
	for rows.Next() {
		var m models.Media
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}

func (r *Repository) CreateMedia(ctx context.Context, m *models.Media) error {
	m.ID = uuid.New()
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt

	_, err := r.DB.ExecContext(ctx, "INSERT INTO media (id, name, type, url, duration_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		m.ID, m.Name, m.Type, m.URL, m.DurationSeconds, m.CreatedAt, m.UpdatedAt)
	return err
}

// --- Playlist Items ---

func (r *Repository) GetPlaylist(ctx context.Context, windowID uuid.UUID) ([]models.PlaylistItem, error) {
	query := `
		SELECT p.id, p.window_id, p.media_id, p.position, p.duration_seconds, p.created_at, p.updated_at,
		       m.id, m.name, m.type, m.url, m.duration_seconds, m.created_at, m.updated_at
		FROM playlist_items p
		JOIN media m ON p.media_id = m.id
		WHERE p.window_id = $1
		ORDER BY p.position ASC
	`
	rows, err := r.DB.QueryContext(ctx, query, windowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PlaylistItem
	for rows.Next() {
		var p models.PlaylistItem
		var m models.Media
		err := rows.Scan(
			&p.ID, &p.WindowID, &p.MediaID, &p.Position, &p.DurationSeconds, &p.CreatedAt, &p.UpdatedAt,
			&m.ID, &m.Name, &m.Type, &m.URL, &m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		p.Media = &m
		items = append(items, p)
	}
	return items, nil
}

func (r *Repository) SetPlaylist(ctx context.Context, windowID uuid.UUID, items []models.PlaylistItem) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing playlist
	_, err = tx.ExecContext(ctx, "DELETE FROM playlist_items WHERE window_id = $1", windowID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for i, item := range items {
		item.ID = uuid.New()
		item.WindowID = windowID
		item.Position = i
		item.CreatedAt = now
		item.UpdatedAt = now
		
		_, err := tx.ExecContext(ctx, "INSERT INTO playlist_items (id, window_id, media_id, position, duration_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			item.ID, item.WindowID, item.MediaID, item.Position, item.DurationSeconds, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// --- Sync Sessions ---

func (r *Repository) CreateSyncSession(ctx context.Context, session *models.SyncSession) error {
	session.ID = uuid.New()
	session.CreatedAt = time.Now().UTC()

	_, err := r.DB.ExecContext(ctx, "INSERT INTO sync_sessions (id, media_id, start_at, end_at, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		session.ID, session.MediaID, session.StartAt, session.EndAt, session.Status, session.CreatedAt)
	return err
}

func (r *Repository) GetActiveOrScheduledSyncSession(ctx context.Context) (*models.SyncSession, error) {
	query := `
		SELECT s.id, s.media_id, s.start_at, s.end_at, s.status, s.created_at,
		       m.id, m.name, m.type, m.url, m.duration_seconds, m.created_at, m.updated_at
		FROM sync_sessions s
		JOIN media m ON s.media_id = m.id
		WHERE s.end_at > $1
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	now := time.Now().UTC()
	row := r.DB.QueryRowContext(ctx, query, now)

	var s models.SyncSession
	var m models.Media
	err := row.Scan(
		&s.ID, &s.MediaID, &s.StartAt, &s.EndAt, &s.Status, &s.CreatedAt,
		&m.ID, &m.Name, &m.Type, &m.URL, &m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No active session
		}
		return nil, err
	}
	s.Media = &m
	return &s, nil
}
