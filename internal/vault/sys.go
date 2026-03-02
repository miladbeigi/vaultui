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

// SealInfo holds seal type and storage backend from /sys/seal-status.
type SealInfo struct {
	SealType    string
	StorageType string
}

// SealStatus queries /sys/seal-status for seal type and storage backend.
func (c *Client) SealStatus() (*SealInfo, error) {
	resp, err := c.raw.Sys().SealStatus()
	if err != nil {
		return nil, fmt.Errorf("vault seal status: %w", err)
	}
	return &SealInfo{
		SealType:    resp.Type,
		StorageType: resp.StorageType,
	}, nil
}

// HAInfo summarises the HA cluster node counts.
type HAInfo struct {
	ActiveNodes  int
	StandbyNodes int
}

// HAStatus queries /sys/ha-status and counts active vs standby nodes.
func (c *Client) HAStatus() (*HAInfo, error) {
	resp, err := c.raw.Sys().HAStatus()
	if err != nil {
		return nil, fmt.Errorf("vault ha status: %w", err)
	}
	info := &HAInfo{}
	for _, n := range resp.Nodes {
		if n.ActiveNode {
			info.ActiveNodes++
		} else {
			info.StandbyNodes++
		}
	}
	return info, nil
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

// ListAuthMethods returns all enabled auth methods, sorted by path.
func (c *Client) ListAuthMethods() ([]MountEntry, error) {
	auths, err := c.raw.Sys().ListAuth()
	if err != nil {
		return nil, fmt.Errorf("listing auth methods: %w", err)
	}

	entries := make([]MountEntry, 0, len(auths))
	for path, auth := range auths {
		entries = append(entries, MountEntry{
			Path:        path,
			Type:        auth.Type,
			Description: auth.Description,
			Accessor:    auth.Accessor,
			Local:       auth.Local,
			SealWrap:    auth.SealWrap,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}

// ListPolicies returns the names of all ACL policies, sorted alphabetically.
func (c *Client) ListPolicies() ([]string, error) {
	policies, err := c.raw.Sys().ListPolicies()
	if err != nil {
		return nil, fmt.Errorf("listing policies: %w", err)
	}
	sort.Strings(policies)
	return policies, nil
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
