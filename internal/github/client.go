package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GitHubClient interface {
	RepoExists(ctx context.Context, repo string) (bool, error)
	GetLatestTag(ctx context.Context, repo string) (string, error)
}

type client struct {
	token string
	http  *http.Client
}

func NewClient(token string) GitHubClient {
	return &client{
		token: token,
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *client) RepoExists(ctx context.Context, repo string) (bool, error) {
	resp, err := c.request(ctx, http.MethodHead, "/repos/"+repo)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (c *client) GetLatestTag(ctx context.Context, repo string) (string, error) {
	var tags []struct{ Name string }
	if err := c.getJSON(ctx, "/repos/"+repo+"/tags?per_page=1", &tags); err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", nil
	}
	return tags[0].Name, nil
}

func (c *client) getJSON(ctx context.Context, path string, out any) error {
	resp, err := c.request(ctx, http.MethodGet, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api error: status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *client) request(ctx context.Context, method, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, "https://api.github.com"+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	for i := 0; i < 3; i++ {
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("execute request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}
		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}
