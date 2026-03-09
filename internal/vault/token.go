package vault

import (
	"fmt"
	"strings"
	"time"
)

// TokenDetails holds comprehensive information about the current token.
type TokenDetails struct {
	Accessor    string
	DisplayName string
	EntityID    string
	Policies    []string
	TokenType   string
	Orphan      bool
	Renewable   bool
	NumUses     int
	TTL         time.Duration
	MaxTTL      time.Duration
	CreationTTL time.Duration
	CreationAt  time.Time
	ExpireAt    time.Time
	Path        string
	Meta        map[string]string
}

// InspectToken returns comprehensive details about the current token.
func (c *Client) InspectToken() (*TokenDetails, error) {
	secret, err := c.raw.Auth().Token().LookupSelf()
	if err != nil {
		return nil, fmt.Errorf("token lookup-self: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("token lookup returned no data")
	}

	d := secret.Data
	td := &TokenDetails{
		Accessor:    getString(d, "accessor"),
		DisplayName: getString(d, "display_name"),
		EntityID:    getString(d, "entity_id"),
		TokenType:   getString(d, "type"),
		Path:        getString(d, "path"),
		Orphan:      getBool(d, "orphan"),
		Renewable:   getBool(d, "renewable"),
		NumUses:     getInt(d, "num_uses"),
		TTL:         getDurationSec(d, "ttl"),
		MaxTTL:      getDurationSec(d, "explicit_max_ttl"),
		CreationTTL: getDurationSec(d, "creation_ttl"),
		Policies:    getStringSlice(d, "policies"),
	}

	if v, ok := d["creation_time"].(float64); ok && v > 0 {
		td.CreationAt = time.Unix(int64(v), 0)
	}

	if v, ok := d["expire_time"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			td.ExpireAt = t
		}
	}

	if meta, ok := d["meta"].(map[string]interface{}); ok {
		td.Meta = make(map[string]string, len(meta))
		for k, v := range meta {
			td.Meta[k] = fmt.Sprintf("%v", v)
		}
	}

	return td, nil
}

// PoliciesString returns a comma-separated list of policies.
func (td *TokenDetails) PoliciesString() string {
	return strings.Join(td.Policies, ", ")
}

func getString(d map[string]interface{}, key string) string {
	if v, ok := d[key].(string); ok {
		return v
	}
	return ""
}

func getBool(d map[string]interface{}, key string) bool {
	if v, ok := d[key].(bool); ok {
		return v
	}
	return false
}

func getInt(d map[string]interface{}, key string) int {
	switch v := d[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func getStringSlice(d map[string]interface{}, key string) []string {
	if v, ok := d[key].([]interface{}); ok {
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func getDurationSec(d map[string]interface{}, key string) time.Duration {
	switch v := d[key].(type) {
	case float64:
		return time.Duration(v) * time.Second
	case int:
		return time.Duration(v) * time.Second
	}
	return 0
}
