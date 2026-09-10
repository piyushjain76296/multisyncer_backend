package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/user/backend_multisyncer/internal/database"
	"github.com/user/backend_multisyncer/internal/handlers"
	"github.com/user/backend_multisyncer/internal/models"
	"github.com/user/backend_multisyncer/internal/repository"
	"github.com/user/backend_multisyncer/internal/websocket"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to PostgreSQL")

	if err := database.Migrate(context.Background(), db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Migrations applied successfully")

	repo := repository.NewRepository(db)
	
	// Seed Data if empty
	seedData(repo)

	hub := websocket.NewHub()
	go hub.Run()

	api := handlers.NewAPI(repo, hub)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on CORS_ORIGIN in prod
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/health", api.Health)
	
	r.Route("/api/windows", func(r chi.Router) {
		r.Get("/", api.GetWindows)
		r.Post("/", api.CreateWindow)
		r.Get("/{id}/playlist", api.GetPlaylist)
		r.Post("/{id}/playlist", api.UpdatePlaylist)
	})

	r.Route("/api/media", func(r chi.Router) {
		r.Get("/", api.GetMedia)
		r.Post("/", api.CreateMedia)
	})

	r.Route("/api/sync", func(r chi.Router) {
		r.Post("/", api.CreateSyncSession)
		r.Get("/active", api.GetActiveSyncSession)
	})

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Starting server on port " + port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func seedData(repo *repository.Repository) {
	ctx := context.Background()
	windows, _ := repo.GetWindows(ctx)
	if len(windows) > 0 {
		return // Already seeded
	}
	
	log.Println("Seeding data...")
	w1, _ := repo.CreateWindow(ctx, "Window 1")
	w2, _ := repo.CreateWindow(ctx, "Window 2")
	w3, _ := repo.CreateWindow(ctx, "Window 3")
	w4, _ := repo.CreateWindow(ctx, "Window 4")

	m1 := &models.Media{Name: "Nature Video", Type: models.MediaTypeVideo, URL: "https://www.w3schools.com/html/mov_bbb.mp4", DurationSeconds: 10}
	m2 := &models.Media{Name: "City Image", Type: models.MediaTypeImage, URL: "https://images.unsplash.com/photo-1477959858617-67f85cf4f1df", DurationSeconds: 5}
	m3 := &models.Media{Name: "Ocean Video", Type: models.MediaTypeVideo, URL: "https://www.w3schools.com/html/mov_bbb.mp4", DurationSeconds: 15}
	m4 := &models.Media{Name: "Mountain Image", Type: models.MediaTypeImage, URL: "https://images.unsplash.com/photo-1464822759023-fed622ff2c3b", DurationSeconds: 5}
	blank := &models.Media{Name: "Blank Screen", Type: models.MediaTypeBlank, URL: "", DurationSeconds: 5}

	repo.CreateMedia(ctx, m1)
	repo.CreateMedia(ctx, m2)
	repo.CreateMedia(ctx, m3)
	repo.CreateMedia(ctx, m4)
	repo.CreateMedia(ctx, blank)

	repo.SetPlaylist(ctx, w1.ID, []models.PlaylistItem{
		{MediaID: m1.ID, DurationSeconds: m1.DurationSeconds},
		{MediaID: m2.ID, DurationSeconds: m2.DurationSeconds},
		{MediaID: m3.ID, DurationSeconds: m3.DurationSeconds},
	})
	
	repo.SetPlaylist(ctx, w2.ID, []models.PlaylistItem{
		{MediaID: m2.ID, DurationSeconds: m2.DurationSeconds},
		{MediaID: m4.ID, DurationSeconds: m4.DurationSeconds},
		{MediaID: blank.ID, DurationSeconds: blank.DurationSeconds},
	})
	
	repo.SetPlaylist(ctx, w3.ID, []models.PlaylistItem{
		{MediaID: m3.ID, DurationSeconds: m3.DurationSeconds},
		{MediaID: m4.ID, DurationSeconds: m4.DurationSeconds},
	})
	
	repo.SetPlaylist(ctx, w4.ID, []models.PlaylistItem{
		{MediaID: m1.ID, DurationSeconds: m1.DurationSeconds},
		{MediaID: m2.ID, DurationSeconds: m2.DurationSeconds},
		{MediaID: m4.ID, DurationSeconds: m4.DurationSeconds},
	})
}
