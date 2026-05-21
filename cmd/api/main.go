package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"ghnotifier/internal/config"
	"ghnotifier/internal/db"
	"ghnotifier/internal/github"
	"ghnotifier/internal/notifier"
	"ghnotifier/internal/scanner"
	"ghnotifier/internal/subscription"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
}

func main() {
	config := config.NewConfig()
	ctx := context.Background()

	pool, err := db.InitPostgres(ctx, config.DBUrl)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	err = db.RunMigrations(config.DBUrl, config.MigrationPath)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	ghClient := github.NewClient(config.GHToken)
	emailNotifier := notifier.NewEmailNotifier(
		config.SMTPHost,
		config.SMTPPort,
		config.SMTPUser,
		config.SMTPPass,
		config.SMTPFrom,
	)
	repo := subscription.NewRepository(pool)
	subService := subscription.NewService(repo, ghClient, emailNotifier)
	subHandler := subscription.NewHandler(subService)

	scanner := scanner.NewScanner(
		repo,
		ghClient,
		emailNotifier,
		3*time.Minute,
	)
	go scanner.Start(ctx)

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		subHandler.RegisterRoutes(r)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}


