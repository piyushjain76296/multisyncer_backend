package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/backend_multisyncer/internal/models"
	"github.com/user/backend_multisyncer/internal/repository"
	"github.com/user/backend_multisyncer/internal/websocket"
)

type API struct {
	Repo *repository.Repository
	Hub  *websocket.Hub
}

func NewAPI(repo *repository.Repository, hub *websocket.Hub) *API {
	return &API{Repo: repo, Hub: hub}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func (a *API) broadcast(eventType string, payload interface{}) {
	msg := websocket.WSMessage{
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Payload:   payload,
	}
	b, _ := json.Marshal(msg)
	a.Hub.Broadcast(b)
}

// Health

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

// Windows

func (a *API) GetWindows(w http.ResponseWriter, r *http.Request) {
	windows, err := a.Repo.GetWindows(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if windows == nil {
		windows = []models.Window{} // prevent null
	}
	respondJSON(w, http.StatusOK, windows)
}

func (a *API) CreateWindow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid name")
		return
	}
	window, err := a.Repo.CreateWindow(r.Context(), req.Name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	a.broadcast("WINDOW_CREATED", window)
	respondJSON(w, http.StatusCreated, window)
}

// Media

func (a *API) GetMedia(w http.ResponseWriter, r *http.Request) {
	media, err := a.Repo.GetMedia(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if media == nil {
		media = []models.Media{}
	}
	respondJSON(w, http.StatusOK, media)
}

func (a *API) CreateMedia(w http.ResponseWriter, r *http.Request) {
	var m models.Media
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid body")
		return
	}
	if m.Name == "" || m.Type == "" || m.DurationSeconds <= 0 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Missing required fields or invalid duration")
		return
	}

	if err := a.Repo.CreateMedia(r.Context(), &m); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	a.broadcast("MEDIA_CREATED", m)
	respondJSON(w, http.StatusCreated, m)
}

// Playlists

func (a *API) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	windowIDStr := chi.URLParam(r, "id")
	windowID, err := uuid.Parse(windowIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid window ID")
		return
	}
	
	items, err := a.Repo.GetPlaylist(r.Context(), windowID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if items == nil {
		items = []models.PlaylistItem{}
	}
	respondJSON(w, http.StatusOK, items)
}

func (a *API) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	windowIDStr := chi.URLParam(r, "id")
	windowID, err := uuid.Parse(windowIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid window ID")
		return
	}

	var items []models.PlaylistItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid body")
		return
	}

	if err := a.Repo.SetPlaylist(r.Context(), windowID, items); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	itemsRefetched, _ := a.Repo.GetPlaylist(r.Context(), windowID)
	a.broadcast("PLAYLIST_UPDATED", map[string]interface{}{
		"window_id": windowID,
		"items": itemsRefetched,
	})
	respondJSON(w, http.StatusOK, itemsRefetched)
}

// Sync

func (a *API) CreateSyncSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MediaID uuid.UUID `json:"media_id"`
		Duration int      `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid body")
		return
	}
	if req.Duration <= 0 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid duration")
		return
	}

	now := time.Now().UTC()
	startAt := now.Add(1 * time.Second) // Small buffer to ensure all clients start precisely together
	endAt := startAt.Add(time.Duration(req.Duration) * time.Second)

	session := &models.SyncSession{
		MediaID: req.MediaID,
		StartAt: startAt,
		EndAt:   endAt,
		Status:  models.SyncStatusScheduled,
	}

	if err := a.Repo.CreateSyncSession(r.Context(), session); err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Fetch full session with media for broadcast
	fullSession, _ := a.Repo.GetActiveOrScheduledSyncSession(r.Context())
	a.broadcast("SYNC_SCHEDULED", fullSession)
	respondJSON(w, http.StatusCreated, fullSession)
}

func (a *API) GetActiveSyncSession(w http.ResponseWriter, r *http.Request) {
	session, err := a.Repo.GetActiveOrScheduledSyncSession(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, session) // could be null
}
