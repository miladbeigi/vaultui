package vault

import (
	"fmt"
	"sort"
)

// PKICert represents a certificate entry from PKI.
type PKICert struct {
	SerialNumber string
}

// PKIRole represents a PKI role.
type PKIRole struct {
	Name string
}

// ListPKICerts lists certificates from a PKI engine mount.
func (c *Client) ListPKICerts(mount string) ([]PKICert, error) {
	cacheKey := "pki:certs:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]PKICert), nil
	}

	secret, err := c.raw.Logical().List(mount + "certs")
	if err != nil {
		return nil, fmt.Errorf("listing PKI certs at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	certs := make([]PKICert, 0, len(keysRaw))
	for _, k := range keysRaw {
		if serial, ok := k.(string); ok {
			certs = append(certs, PKICert{SerialNumber: serial})
		}
	}

	sort.Slice(certs, func(i, j int) bool {
		return certs[i].SerialNumber < certs[j].SerialNumber
	})

	c.cache.Set(cacheKey, certs)
	return certs, nil
}

// ListPKIRoles lists roles from a PKI engine mount.
func (c *Client) ListPKIRoles(mount string) ([]PKIRole, error) {
	cacheKey := "pki:roles:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]PKIRole), nil
	}

	secret, err := c.raw.Logical().List(mount + "roles")
	if err != nil {
		return nil, fmt.Errorf("listing PKI roles at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	roles := make([]PKIRole, 0, len(keysRaw))
	for _, k := range keysRaw {
		if name, ok := k.(string); ok {
			roles = append(roles, PKIRole{Name: name})
		}
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})

	c.cache.Set(cacheKey, roles)
	return roles, nil
}

// PKICertDetail holds the details of a single PKI certificate.
type PKICertDetail struct {
	SerialNumber string
	Certificate  string
	CAChain      string
	Revocation   string
}

// ReadPKICert reads the details of a specific certificate.
func (c *Client) ReadPKICert(mount, serial string) (*PKICertDetail, error) {
	readPath := mount + "cert/" + serial
	secret, err := c.raw.Logical().Read(readPath)
	if err != nil {
		return nil, fmt.Errorf("reading PKI cert %q: %w", readPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no cert found at %q", readPath)
	}

	detail := &PKICertDetail{SerialNumber: serial}
	if v, ok := secret.Data["certificate"].(string); ok {
		detail.Certificate = v
	}
	if v, ok := secret.Data["ca_chain"].(string); ok {
		detail.CAChain = v
	}
	if v, ok := secret.Data["revocation_time"].(string); ok {
		detail.Revocation = v
	}

	return detail, nil
}
