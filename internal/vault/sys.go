package vault

import (
	"fmt"
	"sort"
)

// HealthStatus represents the result of a Vault health check.
type HealthStatus struct {
	Initialized bool
	Sealed      bool
	Standby     bool
	Version     string
	ClusterName string
	ClusterID   string
	Enterprise  bool
}

// Health queries the Vault /sys/health endpoint.
func (c *Client) Health() (*HealthStatus, error) {
	resp, err := c.raw.Sys().Health()
	if err != nil {
		return nil, fmt.Errorf("vault health check: %w", err)
	}

	return &HealthStatus{
		Initialized: resp.Initialized,
		Sealed:      resp.Sealed,
		Standby:     resp.Standby,
		Version:     resp.Version,
		ClusterName: resp.ClusterName,
		ClusterID:   resp.ClusterID,
		Enterprise:  resp.Enterprise,
	}, nil
}

// MountEntry represents a single secret engine or auth method mount.
type MountEntry struct {
	Path           string
	Type           string
	Description    string
	Version        string
	RunningVersion string
	Accessor       string
	Local          bool
	SealWrap       bool
}

// ListSecretEngines returns all mounted secret engines, sorted by path.
func (c *Client) ListSecretEngines() ([]MountEntry, error) {
	mounts, err := c.raw.Sys().ListMounts()
	if err != nil {
		return nil, fmt.Errorf("listing secret engines: %w", err)
	}

	entries := make([]MountEntry, 0, len(mounts))
	for path, mount := range mounts {
		version := mount.Options["version"]
		if version != "" {
			version = "v" + version
		}
		entries = append(entries, MountEntry{
			Path:           path,
			Type:           mount.Type,
			Description:    mount.Description,
			Version:        version,
			RunningVersion: mount.RunningVersion,
			Accessor:       mount.Accessor,
			Local:          mount.Local,
			SealWrap:       mount.SealWrap,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}
