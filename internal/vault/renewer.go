package vault

import (
	"context"
	"fmt"
	"time"
)

// TokenInfo holds the current token's metadata.
type TokenInfo struct {
	TTL       time.Duration
	Renewable bool
	ExpireAt  time.Time
}

// LookupSelfToken returns metadata about the current token.
func (c *Client) LookupSelfToken() (*TokenInfo, error) {
	secret, err := c.raw.Auth().Token().LookupSelf()
	if err != nil {
		return nil, fmt.Errorf("token lookup-self: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("token lookup returned no data")
	}

	ttl := time.Duration(0)
	if v, ok := secret.Data["ttl"]; ok {
		switch t := v.(type) {
		case float64:
			ttl = time.Duration(t) * time.Second
		case int:
			ttl = time.Duration(t) * time.Second
		}
	}

	renewable := false
	if v, ok := secret.Data["renewable"].(bool); ok {
		renewable = v
	}

	return &TokenInfo{
		TTL:       ttl,
		Renewable: renewable,
		ExpireAt:  time.Now().Add(ttl),
	}, nil
}

// RenewSelfToken renews the current token. Returns the new TTL.
func (c *Client) RenewSelfToken() (time.Duration, error) {
	secret, err := c.raw.Auth().Token().RenewSelf(0)
	if err != nil {
		return 0, fmt.Errorf("token renew-self: %w", err)
	}
	if secret == nil || secret.Auth == nil {
		return 0, fmt.Errorf("token renew returned no auth data")
	}

	ttl := time.Duration(secret.Auth.LeaseDuration) * time.Second
	return ttl, nil
}

// TokenRenewer periodically renews the client's token in the background.
type TokenRenewer struct {
	client *Client
	cancel context.CancelFunc
}

// StartTokenRenewer starts background token renewal. It renews at ~2/3
// of the TTL to keep the token alive. Call Stop() to cancel.
func StartTokenRenewer(client *Client) *TokenRenewer {
	ctx, cancel := context.WithCancel(context.Background())
	r := &TokenRenewer{client: client, cancel: cancel}
	go r.loop(ctx)
	return r
}

// Stop cancels the background renewal goroutine.
func (r *TokenRenewer) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *TokenRenewer) loop(ctx context.Context) {
	info, err := r.client.LookupSelfToken()
	if err != nil || !info.Renewable || info.TTL == 0 {
		return
	}

	renewInterval := info.TTL * 2 / 3
	if renewInterval < 10*time.Second {
		renewInterval = 10 * time.Second
	}

	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newTTL, err := r.client.RenewSelfToken()
			if err != nil {
				return
			}
			renewInterval = newTTL * 2 / 3
			if renewInterval < 10*time.Second {
				renewInterval = 10 * time.Second
			}
			ticker.Reset(renewInterval)
		}
	}
}
