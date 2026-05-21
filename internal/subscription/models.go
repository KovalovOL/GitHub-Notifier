package subscription

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Owner       string    `json:"owner"`
	Repo        string    `json:"repo"`
	Token       string    `json:"token"`
	IsConfirmed bool      `json:"is_confirmed"`
	LastTag     string    `json:"last_tag"`
	LastChecked time.Time `json:"last_checked"`
}

type SubscriptionDTO struct {
	RepoPath    string    `json:"repo_path"`
	LastTag     string    `json:"last_tag"`
	IsConfirmed bool      `json:"is_confirmed"`
	LastChecked time.Time `json:"last_checked"`
}