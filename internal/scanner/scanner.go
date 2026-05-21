package scanner

import (
	"context"
	"fmt"
	"log"
	"time"

	"ghnotifier/internal/github"
	"ghnotifier/internal/notifier"
	"ghnotifier/internal/subscription"
)

type Scanner struct {
	repo     subscription.Repository
	ghClient github.GitHubClient
	notifier notifier.Notifier
	interval time.Duration
}

func NewScanner(repo subscription.Repository, ghClient github.GitHubClient, notifier notifier.Notifier, interval time.Duration) *Scanner {
	return &Scanner{
		repo:     repo,
		ghClient: ghClient,
		notifier: notifier,
		interval: interval,
	}
}
 
func (s *Scanner) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	if err := s.scan(ctx); err != nil {
		log.Printf("Scanner error on start: %v", err)
	}


	for {
		select {
		case <-ctx.Done():
			log.Println("Scanner stopped")
			return
		case <-ticker.C:
			log.Println("Scanner started")
			if err := s.scan(ctx); err != nil {
				log.Printf("Scanner error: %v", err)
			}
			log.Println("Scanner completed")
		}
	}
}

func (s *Scanner) scan(ctx context.Context) error {
	subs, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	cache := make(map[string]string)
	for _, sub := range subs {
		if !sub.IsConfirmed {
			continue
		}

		repo := sub.Owner + "/" + sub.Repo
		latest, ok := cache[repo]
		if !ok {
			latest, _ = s.ghClient.GetLatestTag(ctx, repo)
			cache[repo] = latest
		}

		if latest != "" && latest != sub.LastTag {
			_ = s.repo.Update(ctx, sub.Owner, sub.Repo, latest)
			_ = s.notifier.SendRepoUpdate(ctx, sub.Email, repo, latest)
		}
	}

	return nil
}
