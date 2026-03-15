package vault

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// AWSRole represents an AWS role entry.
type AWSRole struct {
	Name           string
	CredentialType string
	PolicyARNs     []string
}

// AWSRoleDetail holds the full configuration for an AWS role.
type AWSRoleDetail struct {
	Name            string
	CredentialTypes []string
	RoleARNs        []string
	PolicyARNs      []string
	PolicyDocument  string
	IAMGroups       []string
	DefaultSTSTTL   string
	MaxSTSTTL       string
	UserPath        string
}

// AWSConfig holds the root configuration for an AWS engine (credentials masked by Vault).
type AWSConfig struct {
	AccessKey   string
	Region      string
	IAMEndpoint string
	STSEndpoint string
	MaxRetries  int
}

// AWSLeaseConfig holds the lease configuration for the AWS engine.
type AWSLeaseConfig struct {
	Lease    string
	LeaseMax string
}

// AWSLease represents an active lease under the AWS engine.
type AWSLease struct {
	LeaseID    string
	TTL        time.Duration
	IssueTime  string
	ExpireTime string
	Renewable  bool
}

// ListAWSRoles lists roles defined in an AWS secrets engine.
func (c *Client) ListAWSRoles(mount string) ([]AWSRole, error) {
	cacheKey := "aws:roles:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]AWSRole), nil
	}

	secret, err := c.raw.Logical().List(mount + "roles")
	if err != nil {
		return nil, fmt.Errorf("listing AWS roles at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	roles := make([]AWSRole, 0, len(keysRaw))
	for _, k := range keysRaw {
		name, ok := k.(string)
		if !ok {
			continue
		}
		role := AWSRole{Name: name}
		if detail, err := c.ReadAWSRole(mount, name); err == nil && detail != nil {
			if len(detail.CredentialTypes) > 0 {
				role.CredentialType = strings.Join(detail.CredentialTypes, ", ")
			}
			role.PolicyARNs = detail.PolicyARNs
		}
		roles = append(roles, role)
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})

	c.cache.Set(cacheKey, roles)
	return roles, nil
}

// ReadAWSRole reads full details for an AWS role.
func (c *Client) ReadAWSRole(mount, name string) (*AWSRoleDetail, error) {
	path := mount + "roles/" + name
	secret, err := c.raw.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading AWS role %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no AWS role found at %q", path)
	}

	d := &AWSRoleDetail{Name: name}
	d.CredentialTypes = extractStringSlice(secret.Data, "credential_type")
	if len(d.CredentialTypes) == 0 {
		if v, ok := secret.Data["credential_type"].(string); ok && v != "" {
			d.CredentialTypes = []string{v}
		}
	}
	d.RoleARNs = extractStringSlice(secret.Data, "role_arns")
	d.PolicyARNs = extractStringSlice(secret.Data, "policy_arns")
	if v, ok := secret.Data["policy_document"].(string); ok {
		d.PolicyDocument = v
	}
	d.IAMGroups = extractStringSlice(secret.Data, "iam_groups")
	d.DefaultSTSTTL = extractDurationField(secret.Data, "default_sts_ttl")
	d.MaxSTSTTL = extractDurationField(secret.Data, "max_sts_ttl")
	if v, ok := secret.Data["user_path"].(string); ok {
		d.UserPath = v
	}
	return d, nil
}

// ReadAWSConfig reads the root configuration for an AWS engine (secret_key is never returned).
func (c *Client) ReadAWSConfig(mount string) (*AWSConfig, error) {
	path := mount + "config/root"
	secret, err := c.raw.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading AWS config at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no AWS config found at %q", path)
	}

	cfg := &AWSConfig{}
	if v, ok := secret.Data["access_key"].(string); ok {
		cfg.AccessKey = v
	}
	if v, ok := secret.Data["region"].(string); ok {
		cfg.Region = v
	}
	if v, ok := secret.Data["iam_endpoint"].(string); ok {
		cfg.IAMEndpoint = v
	}
	if v, ok := secret.Data["sts_endpoint"].(string); ok {
		cfg.STSEndpoint = v
	}
	if v, ok := secret.Data["max_retries"].(float64); ok {
		cfg.MaxRetries = int(v)
	}
	return cfg, nil
}

// ReadAWSLeaseConfig reads the lease configuration for an AWS engine.
func (c *Client) ReadAWSLeaseConfig(mount string) (*AWSLeaseConfig, error) {
	path := mount + "config/lease"
	secret, err := c.raw.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading AWS lease config at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return &AWSLeaseConfig{}, nil
	}

	lc := &AWSLeaseConfig{}
	if v, ok := secret.Data["lease"].(string); ok {
		lc.Lease = v
	}
	if v, ok := secret.Data["lease_max"].(string); ok {
		lc.LeaseMax = v
	}
	return lc, nil
}

// ListAWSLeases lists active leases under the AWS engine mount.
func (c *Client) ListAWSLeases(mount string) ([]AWSLease, error) {
	prefix := "aws/creds"
	if mount != "aws/" {
		prefix = strings.TrimSuffix(mount, "/") + "/creds"
	}

	secret, err := c.raw.Logical().List("sys/leases/lookup/" + prefix)
	if err != nil {
		return nil, fmt.Errorf("listing AWS leases: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	var leases []AWSLease
	for _, k := range keysRaw {
		roleName, ok := k.(string)
		if !ok {
			continue
		}
		roleName = strings.TrimSuffix(roleName, "/")
		roleLeases, err := c.listAWSRoleLeases(prefix + "/" + roleName)
		if err != nil {
			continue
		}
		leases = append(leases, roleLeases...)
	}

	sort.Slice(leases, func(i, j int) bool {
		return leases[i].LeaseID < leases[j].LeaseID
	})

	return leases, nil
}

func (c *Client) listAWSRoleLeases(prefix string) ([]AWSLease, error) {
	secret, err := c.raw.Logical().List("sys/leases/lookup/" + prefix)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	leases := make([]AWSLease, 0, len(keysRaw))
	for _, k := range keysRaw {
		leaseID, ok := k.(string)
		if !ok {
			continue
		}
		fullID := prefix + "/" + strings.TrimSuffix(leaseID, "/")
		lease := AWSLease{LeaseID: fullID}

		if info, err := c.raw.Logical().Write("sys/leases/lookup", map[string]interface{}{
			"lease_id": fullID,
		}); err == nil && info != nil && info.Data != nil {
			if v, ok := info.Data["ttl"].(float64); ok {
				lease.TTL = time.Duration(v) * time.Second
			}
			if v, ok := info.Data["issue_time"].(string); ok {
				lease.IssueTime = v
			}
			if v, ok := info.Data["expire_time"].(string); ok {
				lease.ExpireTime = v
			}
			if v, ok := info.Data["renewable"].(bool); ok {
				lease.Renewable = v
			}
		}

		leases = append(leases, lease)
	}

	return leases, nil
}
