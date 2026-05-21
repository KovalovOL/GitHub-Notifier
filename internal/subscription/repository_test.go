package subscription

import (
	"context"
	"testing"

	"ghnotifier/internal/config"
	"ghnotifier/internal/db"
	"github.com/stretchr/testify/require"
)

func TestRepositoryIntegration(t *testing.T) {
	cfg := config.NewConfig()

	ctx := context.Background()
	pool, err := db.InitPostgres(ctx, cfg.DBUrl)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	cleanup := func() {
		_, err := pool.Exec(ctx, "DELETE FROM subscriptions")
		require.NoError(t, err)
	}

	t.Run("Create", func(t *testing.T) {
		cleanup()
		token, err := repo.Create(ctx, "test@example.com", "google", "go", "v1.0.0")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Test uniqueness on email, owner, repo
		_, err = repo.Create(ctx, "test@example.com", "google", "go", "v1.0.0")
		require.Error(t, err) // now we expect an error on duplicate

		subs, err := repo.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		require.Len(t, subs, 1)
		require.Equal(t, "test@example.com", subs[0].Email)
		require.Equal(t, "google", subs[0].Owner)
		require.Equal(t, "go", subs[0].Repo)
		require.Equal(t, token, subs[0].Token)
		require.False(t, subs[0].IsConfirmed)
	})

	t.Run("Confirm and GetByToken", func(t *testing.T) {
		cleanup()
		token, err := repo.Create(ctx, "test@example.com", "apple", "swift", "v5.9")
		require.NoError(t, err)

		sub, err := repo.GetByToken(ctx, token)
		require.NoError(t, err)
		require.False(t, sub.IsConfirmed)

		err = repo.Confirm(ctx, token)
		require.NoError(t, err)

		sub, err = repo.GetByToken(ctx, token)
		require.NoError(t, err)
		require.True(t, sub.IsConfirmed)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		cleanup()
		_, _ = repo.Create(ctx, "user1@example.com", "owner1", "repo-a", "")
		_, _ = repo.Create(ctx, "user2@example.com", "owner2", "repo-b", "")
		_, _ = repo.Create(ctx, "user1@example.com", "owner3", "repo-c", "")

		subs, err := repo.GetByEmail(ctx, "user1@example.com")
		require.NoError(t, err)
		require.Len(t, subs, 2)
		for _, s := range subs {
			require.Equal(t, "user1@example.com", s.Email)
		}
	})

	t.Run("Update", func(t *testing.T) {
		cleanup()
		_, _ = repo.Create(ctx, "test@example.com", "apple", "swift", "")
		
		err := repo.Update(ctx, "apple", "swift", "v5.9")
		require.NoError(t, err)

		subs, err := repo.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		require.Len(t, subs, 1)
		require.Equal(t, "v5.9", subs[0].LastTag)
		require.NotZero(t, subs[0].LastChecked)
	})

	t.Run("Delete and DeleteByToken", func(t *testing.T) {
		cleanup()
		token, err := repo.Create(ctx, "test@example.com", "microsoft", "vscode", "")
		require.NoError(t, err)

		err = repo.DeleteByToken(ctx, token)
		require.NoError(t, err)

		subs, err := repo.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		require.Empty(t, subs)
		
		// Test Delete (original)
		newToken, _ := repo.Create(ctx, "test@example.com", "microsoft", "vscode", "")
		err = repo.DeleteByToken(ctx, newToken)
		require.NoError(t, err)
		
		subs, _ = repo.GetByEmail(ctx, "test@example.com")
		require.Empty(t, subs)
	})
}
