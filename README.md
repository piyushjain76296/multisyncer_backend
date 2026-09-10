<div align="center">
  <h1>🎬 Multi-Window Media Sequencer</h1>
  <p><strong>A production-grade, highly deterministic real-time media synchronization engine.</strong></p>

  [![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](#)
  [![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?style=for-the-badge&logo=typescript&logoColor=white)](#)
  [![Vite](https://img.shields.io/badge/Vite-B73BFE?style=for-the-badge&logo=vite&logoColor=FFD62E)](#)
  [![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#)
  [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)](#)

</div>

---

## 📖 Overview

The **Multi-Window Media Sequencer** is a powerful full-stack application designed to orchestrate independent media sequences across multiple browser windows simultaneously. It features a custom **stateless deterministic playback engine** that ensures all screens remain in perfect lockstep using a highly efficient server-authoritative time synchronization protocol over WebSockets.

Whether displaying digital signage, multi-screen art installations, or synchronized presentations, this system ensures frame-perfect alignment without relying on clunky state machines or constant server polling.

---

## ✨ Key Features

- **⏱️ Deterministic Synchronization:** Employs a 5-hour logical epoch to mathematically calculate the exact frame every window should be displaying based purely on synchronized server time. 
- **🌐 Real-Time WebSockets:** Instantly broadcasts state changes (new media, updated playlists, scheduled syncs) to all connected clients.
- **🎛️ Full Control Panel:** A comprehensive, sleek React frontend to manage media libraries, spin up virtual windows, and dynamically edit drag-and-drop style playlists on the fly.
- **⚡ "Global Sync" Override:** Instantly hijack all screens to play a specific media asset simultaneously, gracefully returning them to their exact logical positions in their individual playlists when finished.
- **🖼️ Rich Media Support:** Natively handles Video URLs, static Images, and Blank placeholders.

---

## 🏗️ Architecture

The platform follows a modern separated architecture:

### Frontend (`/frontend`)
*   **Tech Stack:** React, TypeScript, Vite, CSS Modules.
*   **Design Pattern:** Completely stateless client. It never "guesses" what comes next; it simply calculates its state (`currentMedia`, `mediaStartTimeMs`) based on the globally synchronized clock and its assigned playlist. 

### Backend (`/backend`)
*   **Tech Stack:** Go 1.21+, `go-chi` (Routing), `gorilla/websocket` (Real-time).
*   **Design Pattern:** Acts as the absolute source of truth. Handles CRUD operations for the PostgreSQL database and maintains the central time clock for the entire fleet of screens.

---

## 🚀 Getting Started

Follow these instructions to run the project locally.

### Prerequisites
*   [Docker](https://www.docker.com/) & Docker Compose
*   [Go 1.21+](https://go.dev/)
*   [Node.js 18+](https://nodejs.org/)

### 1. Database (PostgreSQL)
Start the local database using Docker:
```bash
docker-compose up -d
```
*(This exposes PostgreSQL on `localhost:5432`)*

### 2. Backend Server
Navigate to the backend directory, install dependencies, and start the Go server:
```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server/main.go
```
*(The server automatically runs database migrations, seeds demo data, and listens on port `8080`)*

### 3. Frontend Dashboard
In a new terminal, start the Vite development server:
```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```
Navigate to **[http://localhost:5173](http://localhost:5173)** in your browser to access the Control Panel.

---

## 🧠 How the Engine Works

Instead of the server frantically pushing "play next" events to clients every few seconds (which leads to drift and lag), we use **Math**.

1.  **Logical Epoch**: A continuous 5-hour cycle (18,000 seconds) is established globally.
2.  **Calculation**: At any given animation frame, the client calculates `timeInCycle = serverTime % 18000`.
3.  **Playlist Wrap**: By dividing `timeInCycle` by the total duration of a window's playlist, the engine mathematically determines exactly which media item should be active at this exact millisecond, and exactly what its video `currentTime` should be.
4.  **Resilience**: If a client disconnects, buffers, or refreshes, it doesn't matter. The moment it reconnects, the math drops it perfectly back into sync with the rest of the fleet.

---

## 🔌 API Reference

### REST Endpoints
*   `GET /api/windows` - List all screens.
*   `POST /api/windows` - Provision a new screen.
*   `GET /api/windows/:id/playlist` - Retrieve a screen's playlist.
*   `POST /api/windows/:id/playlist` - Update a screen's playlist.
*   `GET /api/media` - List global media library.
*   `POST /api/media` - Add new media to library.
*   `POST /api/sync` - Schedule a global sync event.
*   `GET /api/sync/active` - Get the currently active/scheduled sync.

### WebSocket Protocol
*   **Endpoint**: `ws://<host>/ws`
*   **Events**: `TIME_SYNC`, `WINDOW_CREATED`, `MEDIA_CREATED`, `PLAYLIST_UPDATED`, `SYNC_SCHEDULED`.

---

## 📦 Deployment Strategy

This application is production-ready and can be deployed easily:

*   **Database**: Deploy to a managed PostgreSQL provider (e.g., [Supabase](https://supabase.com/), [Neon](https://neon.tech/)).
*   **Backend**: Package via Dockerfile and deploy to [Render](https://render.com/) or [Railway](https://railway.app/). Ensure `DATABASE_URL` and `CORS_ORIGIN` environment variables are set.
*   **Frontend**: Build the static React bundle (`npm run build`) and deploy to [Vercel](https://vercel.com/) or [Netlify](https://www.netlify.com/). Set `VITE_API_URL` and `VITE_WS_URL`.

---

<div align="center">
  <i>Built with ❤️ for perfectly synchronized pixels.</i>
</div>
