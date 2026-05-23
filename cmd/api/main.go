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

	httpSwagger "github.com/swaggo/http-swagger"
	_ "ghnotifier/docs"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
}

// @title          GitHub Notifier API
// @version        1.0
// @description    GitHub Notifier is a service that monitors GitHub repositories and sends email notifications when new commits are pushed to the repositories.
// @termsOfService http://swagger.io/terms/

// @contact.name   Oleksandr
// @contact.email  sasha.kovalov2008@gmail.com

// @host      gh-notifier.online
// @BasePath  /api

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
	emailNotifier := notifier.NewAsyncNotifier(notifier.NewEmailNotifier(
		config.SMTPHost,
		config.SMTPPort,
		config.SMTPUser,
		config.SMTPPass,
		config.SMTPFrom,
	))
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

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}


