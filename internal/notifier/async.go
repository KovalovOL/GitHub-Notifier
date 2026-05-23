package notifier

import (
	"context"
	"log"
)

type asyncNotifier struct {
	base Notifier
}

func NewAsyncNotifier(base Notifier) Notifier {
	return &asyncNotifier{base: base}
}

func (a *asyncNotifier) SendVerificationCode(ctx context.Context, email string, token string) error {
	go func() {
		if err := a.base.SendVerificationCode(context.Background(), email, token); err != nil {
			log.Printf("[AsyncNotifier] Failed to send verification code to %s: %v", email, err)
		}
	}()
	return nil
}

func (a *asyncNotifier) SendRepoUpdate(ctx context.Context, email string, repoName string, newTag string) error {
	go func() {
		if err := a.base.SendRepoUpdate(context.Background(), email, repoName, newTag); err != nil {
			log.Printf("[AsyncNotifier] Failed to send repo update to %s for %s: %v", email, repoName, err)
		}
	}()
	return nil
}

func (a *asyncNotifier) SendSubscriptionSuccess(ctx context.Context, email string, repoName string) error {
	go func() {
		if err := a.base.SendSubscriptionSuccess(context.Background(), email, repoName); err != nil {
			log.Printf("[AsyncNotifier] Failed to send subscription success to %s for %s: %v", email, repoName, err)
		}
	}()
	return nil
}
