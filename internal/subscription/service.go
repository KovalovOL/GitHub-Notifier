package subscription

import (
	"context"
	"fmt"
	"strings"

	"ghnotifier/internal/github"
	"ghnotifier/internal/notifier"
)

type Service interface {
	Subscribe(ctx context.Context, email, repoPath string) error
	ConfirmSubscription(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
	GetAll(ctx context.Context, email string) ([]SubscriptionDTO, error)
}

type service struct {
	repo     Repository
	ghClient github.GitHubClient
	notifier notifier.Notifier
}

func NewService(repo Repository, ghClient github.GitHubClient, notifier notifier.Notifier) Service {
	return &service{
		repo:     repo,
		ghClient: ghClient,
		notifier: notifier,
	}
}

func parseRepoPath(repoPath string) (owner, repo string, err error) {
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository path format, expected 'owner/repo'")
	}
	return parts[0], parts[1], nil
}

func (s *service) Subscribe(ctx context.Context, email, repoPath string) error {
	owner, repo, err := parseRepoPath(repoPath)
	if err != nil {
		return err
	}

	exists, err := s.ghClient.RepoExists(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to check if repository exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("repository %s not found on GitHub", repoPath)
	}

	tag, err := s.ghClient.GetLatestTag(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}

	token, err := s.createSubscription(ctx, email, owner, repo, tag)
	if err != nil {
		return err
	}

	if err := s.notifier.SendVerificationCode(ctx, email, token); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func (s *service) createSubscription(ctx context.Context, email, owner, repo, tag string) (string, error) {
	isSubscribed, err := s.repo.Exists(ctx, email, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to check subscription status: %w", err)
	}
	if isSubscribed {
		return "", fmt.Errorf("you are already subscribed to this repository")
	}

	token, err := s.repo.Create(ctx, email, owner, repo, tag)
	if err != nil {
		return "", fmt.Errorf("failed to create subscription: %w", err)
	}

	return token, nil
}

func (s *service) ConfirmSubscription(ctx context.Context, token string) error {
	sub, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired confirmation token: %w", err)
	}

	if err := s.repo.Confirm(ctx, token); err != nil {
		return fmt.Errorf("failed to confirm subscription: %w", err)
	}

	if err := s.notifier.SendSubscriptionSuccess(ctx, sub.Email, sub.Owner+"/"+sub.Repo); err != nil {
		return fmt.Errorf("failed to send notification email: %w", err)
	}

	return nil
}

func (s *service) Unsubscribe(ctx context.Context, token string) error {
	_, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid subscription token: %w", err)
	}

	if err := s.repo.DeleteByToken(ctx, token); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	return nil
}

func (s *service) GetAll(ctx context.Context, email string) ([]SubscriptionDTO, error) {
	subs, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	var dtos []SubscriptionDTO
	for _, sub := range subs {
		dtos = append(dtos, SubscriptionDTO{
			RepoPath:    fmt.Sprintf("%s/%s", sub.Owner, sub.Repo),
			LastTag:     sub.LastTag,
			IsConfirmed: sub.IsConfirmed,
			LastChecked: sub.LastChecked,
		})
	}

	return dtos, nil
}
