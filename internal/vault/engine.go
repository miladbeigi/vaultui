package vault

import (
	"fmt"
	"strings"
	"time"
)

// EngineConfig holds detailed mount configuration for a secret engine.
type EngineConfig struct {
	Path                      string
	Type                      string
	Description               string
	UUID                      string
	Accessor                  string
	PluginVersion             string
	RunningVersion            string
	Local                     bool
	SealWrap                  bool
	ExternalEntropyAccess     bool
	DefaultLeaseTTL           time.Duration
	MaxLeaseTTL               time.Duration
	ForceNoCache              bool
	ListingVisibility         string
	TokenType                 string
	Options                   map[string]string
	AuditNonHMACRequestKeys   []string
	AuditNonHMACResponseKeys  []string
	PassthroughRequestHeaders []string
}

// ReadEngineConfig reads mount configuration for a secret engine at the given path.
func (c *Client) ReadEngineConfig(path string) (*EngineConfig, error) {
	cacheKey := "sys/mounts/" + path
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.(*EngineConfig), nil
	}

	trimmed := strings.TrimSuffix(path, "/")
	mount, err := c.raw.Sys().GetMount(trimmed)
	if err != nil {
		return nil, fmt.Errorf("reading mount config for %q: %w", path, err)
	}

	cfg := &EngineConfig{
		Path:                      path,
		Type:                      mount.Type,
		Description:               mount.Description,
		UUID:                      mount.UUID,
		Accessor:                  mount.Accessor,
		PluginVersion:             mount.PluginVersion,
		RunningVersion:            mount.RunningVersion,
		Local:                     mount.Local,
		SealWrap:                  mount.SealWrap,
		ExternalEntropyAccess:     mount.ExternalEntropyAccess,
		DefaultLeaseTTL:           time.Duration(mount.Config.DefaultLeaseTTL) * time.Second,
		MaxLeaseTTL:               time.Duration(mount.Config.MaxLeaseTTL) * time.Second,
		ForceNoCache:              mount.Config.ForceNoCache,
		ListingVisibility:         mount.Config.ListingVisibility,
		TokenType:                 mount.Config.TokenType,
		Options:                   mount.Options,
		AuditNonHMACRequestKeys:   mount.Config.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  mount.Config.AuditNonHMACResponseKeys,
		PassthroughRequestHeaders: mount.Config.PassthroughRequestHeaders,
	}

	c.cache.Set(cacheKey, cfg)
	return cfg, nil
}
