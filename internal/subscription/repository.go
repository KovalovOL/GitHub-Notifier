package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetAll(ctx context.Context) ([]Subscription, error)
	GetByEmail(ctx context.Context, email string) ([]Subscription, error)
	GetByToken(ctx context.Context, token string) (*Subscription, error)
	Exists(ctx context.Context, email, owner, repo string) (bool, error)
	Create(ctx context.Context, email, owner, repo, tag string) (string, error)
	Confirm(ctx context.Context, token string) error
	Update(ctx context.Context, owner, repo string, lastTag string) error
	DeleteByToken(ctx context.Context, token string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetAll(ctx context.Context) ([]Subscription, error) {
	query := `SELECT id, email, owner, repo, token, is_confirmed, COALESCE(last_tag, ''), last_checked FROM subscriptions`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.Email, &s.Owner, &s.Repo, &s.Token, &s.IsConfirmed, &s.LastTag, &s.LastChecked); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, s)
	}

	return subs, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) ([]Subscription, error) {
	query := `SELECT id, email, owner, repo, token, is_confirmed, COALESCE(last_tag, ''), last_checked FROM subscriptions WHERE email = $1`
	rows, err := r.pool.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions by email: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.Email, &s.Owner, &s.Repo, &s.Token, &s.IsConfirmed, &s.LastTag, &s.LastChecked); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, s)
	}

	return subs, nil
}

func (r *repository) GetByToken(ctx context.Context, token string) (*Subscription, error) {
	query := `SELECT id, email, owner, repo, token, is_confirmed, COALESCE(last_tag, ''), last_checked FROM subscriptions WHERE token = $1`
	var s Subscription
	err := r.pool.QueryRow(ctx, query, token).Scan(&s.ID, &s.Email, &s.Owner, &s.Repo, &s.Token, &s.IsConfirmed, &s.LastTag, &s.LastChecked)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("query subscription by token: %w", err)
	}

	return &s, nil
}

func (r *repository) Exists(ctx context.Context, email, owner, repo string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = $1 AND owner = $2 AND repo = $3)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email, owner, repo).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check if subscription exists: %w", err)
	}
	return exists, nil
}

func (r *repository) Create(ctx context.Context, email, owner, repo, tag string) (string, error) {
	query := `INSERT INTO subscriptions (email, owner, repo, last_tag) VALUES ($1, $2, $3, $4) RETURNING token`
	var token string
	err := r.pool.QueryRow(ctx, query, email, owner, repo, tag).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("create subscription: %w", err)
	}
	return token, nil
}

func (r *repository) Confirm(ctx context.Context, token string) error {
	query := `UPDATE subscriptions SET is_confirmed = TRUE WHERE token = $1`
	tag, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("confirm subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, owner, repo string, lastTag string) error {
	query := `UPDATE subscriptions SET last_tag = $3, last_checked = $4 WHERE owner = $1 AND repo = $2`
	_, err := r.pool.Exec(ctx, query, owner, repo, lastTag, time.Now())
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}

func (r *repository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM subscriptions WHERE token = $1`
	tag, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete subscription by token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}
