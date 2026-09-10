# Multi-Window Media Sequencer

A production-quality full-stack application that continuously plays independent media sequences across multiple windows with deterministic, server-authoritative synchronization.

## Architecture

This application consists of:
*   **Frontend**: React, TypeScript, Vite. (Stateless deterministic playback engine).
*   **Backend**: Go, `go-chi`, `gorilla/websocket`. (Source of truth for configuration and time).
*   **Database**: PostgreSQL. (Persistent configuration storage).

The frontend and backend communicate via **REST** for configuration mutations, and **WebSockets** for real-time state broadcasts and time synchronization.

## The 5-Hour Cycle & Deterministic Playback

This system utilizes a highly deterministic, stateless playback engine on the frontend rather than relying on the server to push "next media" events every few seconds.

1.  **Logical Epoch**: A 5-hour cycle (18,000 seconds) is established.
2.  **Calculation**: At any given frame, the client retrieves the current synchronized server time. It calculates `timeInCycle = serverTime % 18000`.
3.  **Playlist Wrap**: By dividing `timeInCycle` by the total duration of a window's playlist, we find exactly which media item should be active at this exact millisecond.
4.  **Result**: When the 5-hour cycle boundary is reached, the math naturally wraps back to 0, causing the playlist to perfectly restart from its configured starting point.

## Synchronization Architecture

When a Sync is scheduled (e.g., `SYNC M2` for 30 seconds):
1.  The backend writes a `SyncSession` to the database with a future `startAt` and `endAt` timestamp, acting as the absolute source of truth.
2.  This session is broadcasted to all connected clients over WebSockets.
3.  Every client continually calculates its current frame against the Server Time Offset.
4.  Once `serverTime >= startAt`, every client simultaneously overrides its normal playlist and begins playing the Sync Media.
5.  Once `serverTime >= endAt`, the Sync logic inherently yields. The frontend playback engine immediately falls back to calculating its position in the normal 5-hour cycle. 
6.  **Crucially, normal playlists are never destroyed or paused.** The sync merely overlays the normal timeline. The windows "resume" their playlists at the exact logical position they *would have been in* had the sync never occurred.

## Setup & Running Locally

### Prerequisites
*   Docker & Docker Compose (for PostgreSQL)
*   Go 1.21+
*   Node.js 18+

### 1. Database
```bash
docker-compose up -d
```
This spins up PostgreSQL on `localhost:5432`.

### 2. Backend
```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server/main.go
```
The backend automatically runs migrations and seeds the database with 4 Windows and Demo Media on startup. It listens on `:8080`.

### 3. Frontend
```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```
The frontend will start on `http://localhost:5173`.

## API Documentation

*   `GET /api/windows`: List all windows.
*   `POST /api/windows`: Create a window.
*   `GET /api/windows/:id/playlist`: Retrieve a window's playlist.
*   `POST /api/windows/:id/playlist`: Update a window's playlist (Provide array of items).
*   `GET /api/media`: List all media items.
*   `POST /api/media`: Create a media item.
*   `POST /api/sync`: Schedule a global sync (`media_id`, `duration_seconds`).
*   `GET /api/sync/active`: Get currently active or scheduled sync.

## WebSocket Protocol

*   **Endpoint**: `ws://<host>/ws`
*   **Message Format**: `{ "type": string, "timestamp": string, "payload": any }`
*   **Events Received**: `TIME_SYNC`, `WINDOW_CREATED`, `MEDIA_CREATED`, `PLAYLIST_UPDATED`, `SYNC_SCHEDULED`.

## Assumptions & Tradeoffs

1.  **Authoritative Duration**: For the 5-hour cycle timeline to be globally deterministic and mathematically sound, the **configured duration** in the playlist is treated as authoritative, rather than the natural video duration. This guarantees that all windows stay in perfect mathematical lockstep regardless of video buffering or metadata loading times.
2.  **Stateless Engine**: Because the frontend calculates state purely from `ServerTime` + `Playlist`, we avoid complex distributed state machines. If a client disconnects and reconnects, it simply resumes rendering whatever the math dictates for the current server time.

## Deployment Strategy

*   **Database**: Deploy to a managed PostgreSQL provider (e.g., Supabase, Neon, Render).
*   **Backend**: Deploy as a web service via Dockerfile on Render or Railway. Set `DATABASE_URL` and `CORS_ORIGIN`.
*   **Frontend**: Build the static React bundle (`npm run build`) and deploy to Vercel or Netlify. Set `VITE_API_URL` and `VITE_WS_URL`.
