package models

import (
	"time"

	"github.com/google/uuid"
)

type Window struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MediaType string

const (
	MediaTypeImage MediaType = "IMAGE"
	MediaTypeVideo MediaType = "VIDEO"
	MediaTypeBlank MediaType = "BLANK"
)

type Media struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Type            MediaType `json:"type"`
	URL             string    `json:"url"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PlaylistItem struct {
	ID              uuid.UUID `json:"id"`
	WindowID        uuid.UUID `json:"window_id"`
	MediaID         uuid.UUID `json:"media_id"`
	Position        int       `json:"position"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	
	// Hydrated field
	Media *Media `json:"media,omitempty"`
}

type SyncStatus string

const (
	SyncStatusScheduled SyncStatus = "SCHEDULED"
	SyncStatusActive    SyncStatus = "ACTIVE"
	SyncStatusCompleted SyncStatus = "COMPLETED"
)

type SyncSession struct {
	ID        uuid.UUID  `json:"id"`
	MediaID   uuid.UUID  `json:"media_id"`
	StartAt   time.Time  `json:"start_at"`
	EndAt     time.Time  `json:"end_at"`
	Status    SyncStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`

	// Hydrated field
	Media *Media `json:"media,omitempty"`
}
