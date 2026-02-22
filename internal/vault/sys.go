package vault

import "fmt"

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
