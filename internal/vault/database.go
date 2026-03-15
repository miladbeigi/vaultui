package vault

import (
	"fmt"
	"sort"
	"strings"
)

// DBConnection represents a database connection configuration entry.
type DBConnection struct {
	Name         string
	PluginName   string
	AllowedRoles []string
}

// DBConnectionDetail holds full configuration for a database connection.
type DBConnectionDetail struct {
	Name                   string
	PluginName             string
	ConnectionURL          string
	AllowedRoles           []string
	RootRotationStatements []string
	VerifyConnection       bool
	PasswordPolicy         string
}

// DBRole represents a dynamic database role.
type DBRole struct {
	Name       string
	DBName     string
	DefaultTTL string
	MaxTTL     string
}

// DBRoleDetail holds full configuration for a database role.
type DBRoleDetail struct {
	Name                 string
	DBName               string
	DefaultTTL           string
	MaxTTL               string
	CreationStatements   []string
	RevocationStatements []string
	RollbackStatements   []string
	RenewStatements      []string
	RoleType             string
}

// DBStaticRole represents a static database role.
type DBStaticRole struct {
	Name           string
	DBName         string
	RotationPeriod string
	Username       string
}

// DBStaticRoleDetail holds full configuration for a static role.
type DBStaticRoleDetail struct {
	Name               string
	DBName             string
	Username           string
	RotationPeriod     string
	LastVaultRotation  string
	RotationStatements []string
}

// ListDBConnections lists database connections at the given mount.
func (c *Client) ListDBConnections(mount string) ([]DBConnection, error) {
	cacheKey := "db:conn:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]DBConnection), nil
	}

	secret, err := c.raw.Logical().List(mount + "config")
	if err != nil {
		return nil, fmt.Errorf("listing db connections at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	conns := make([]DBConnection, 0, len(keysRaw))
	for _, k := range keysRaw {
		name, ok := k.(string)
		if !ok {
			continue
		}
		conn := DBConnection{Name: name}
		if detail, err := c.readDBConnectionSummary(mount, name); err == nil && detail != nil {
			conn.PluginName = detail.PluginName
			conn.AllowedRoles = detail.AllowedRoles
		}
		conns = append(conns, conn)
	}

	sort.Slice(conns, func(i, j int) bool {
		return conns[i].Name < conns[j].Name
	})

	c.cache.Set(cacheKey, conns)
	return conns, nil
}

func (c *Client) readDBConnectionSummary(mount, name string) (*DBConnection, error) {
	secret, err := c.raw.Logical().Read(mount + "config/" + name)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}
	conn := &DBConnection{Name: name}
	if v, ok := secret.Data["plugin_name"].(string); ok {
		conn.PluginName = v
	}
	conn.AllowedRoles = extractStringSlice(secret.Data, "allowed_roles")
	return conn, nil
}

// ReadDBConnection reads full details for a database connection.
func (c *Client) ReadDBConnection(mount, name string) (*DBConnectionDetail, error) {
	path := mount + "config/" + name
	secret, err := c.raw.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading db connection %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no db connection found at %q", path)
	}

	d := &DBConnectionDetail{Name: name}
	if v, ok := secret.Data["plugin_name"].(string); ok {
		d.PluginName = v
	}
	if v, ok := secret.Data["connection_details"].(map[string]interface{}); ok {
		if url, ok := v["connection_url"].(string); ok {
			d.ConnectionURL = url
		}
	}
	d.AllowedRoles = extractStringSlice(secret.Data, "allowed_roles")
	d.RootRotationStatements = extractStringSlice(secret.Data, "root_credentials_rotate_statements")
	if v, ok := secret.Data["verify_connection"].(bool); ok {
		d.VerifyConnection = v
	}
	if v, ok := secret.Data["password_policy"].(string); ok {
		d.PasswordPolicy = v
	}
	return d, nil
}

// ListDBRoles lists dynamic database roles at the given mount.
func (c *Client) ListDBRoles(mount string) ([]DBRole, error) {
	cacheKey := "db:roles:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]DBRole), nil
	}

	secret, err := c.raw.Logical().List(mount + "roles")
	if err != nil {
		return nil, fmt.Errorf("listing db roles at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	roles := make([]DBRole, 0, len(keysRaw))
	for _, k := range keysRaw {
		name, ok := k.(string)
		if !ok {
			continue
		}
		role := DBRole{Name: name}
		if detail, err := c.ReadDBRole(mount, name); err == nil && detail != nil {
			role.DBName = detail.DBName
			role.DefaultTTL = detail.DefaultTTL
			role.MaxTTL = detail.MaxTTL
		}
		roles = append(roles, role)
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})

	c.cache.Set(cacheKey, roles)
	return roles, nil
}

// ReadDBRole reads full details for a dynamic database role.
func (c *Client) ReadDBRole(mount, name string) (*DBRoleDetail, error) {
	path := mount + "roles/" + name
	secret, err := c.raw.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading db role %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no db role found at %q", path)
	}

	d := &DBRoleDetail{Name: name}
	if v, ok := secret.Data["db_name"].(string); ok {
		d.DBName = v
	}
	d.DefaultTTL = extractDurationField(secret.Data, "default_ttl")
	d.MaxTTL = extractDurationField(secret.Data, "max_ttl")
	d.CreationStatements = extractStringSlice(secret.Data, "creation_statements")
	d.RevocationStatements = extractStringSlice(secret.Data, "revocation_statements")
	d.RollbackStatements = extractStringSlice(secret.Data, "rollback_statements")
	d.RenewStatements = extractStringSlice(secret.Data, "renew_statements")
	if v, ok := secret.Data["credential_type"].(string); ok {
		d.RoleType = v
	}
	return d, nil
}

// ListDBStaticRoles lists static database roles at the given mount.
func (c *Client) ListDBStaticRoles(mount string) ([]DBStaticRole, error) {
	cacheKey := "db:static:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]DBStaticRole), nil
	}

	secret, err := c.raw.Logical().List(mount + "static-roles")
	if err != nil {
		return nil, fmt.Errorf("listing db static roles at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	roles := make([]DBStaticRole, 0, len(keysRaw))
	for _, k := range keysRaw {
		name, ok := k.(string)
		if !ok {
			continue
		}
		role := DBStaticRole{Name: name}
		if detail, err := c.ReadDBStaticRole(mount, name); err == nil && detail != nil {
			role.DBName = detail.DBName
			role.RotationPeriod = detail.RotationPeriod
			role.Username = detail.Username
		}
		roles = append(roles, role)
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})

	c.cache.Set(cacheKey, roles)
	return roles, nil
}

// ReadDBStaticRole reads full details for a static database role.
func (c *Client) ReadDBStaticRole(mount, name string) (*DBStaticRoleDetail, error) {
	path := mount + "static-roles/" + name
	secret, err := c.raw.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading db static role %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no db static role found at %q", path)
	}

	d := &DBStaticRoleDetail{Name: name}
	if v, ok := secret.Data["db_name"].(string); ok {
		d.DBName = v
	}
	if v, ok := secret.Data["username"].(string); ok {
		d.Username = v
	}
	d.RotationPeriod = extractDurationField(secret.Data, "rotation_period")
	if v, ok := secret.Data["last_vault_rotation"].(string); ok {
		d.LastVaultRotation = v
	}
	d.RotationStatements = extractStringSlice(secret.Data, "rotation_statements")
	return d, nil
}

// extractStringSlice extracts a []string from a map field that may be []interface{}.
func extractStringSlice(data map[string]interface{}, key string) []string {
	raw, ok := data[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return v
	case string:
		if v != "" {
			return strings.Split(v, ",")
		}
	}
	return nil
}

// extractDurationField extracts a TTL field which may be a json.Number/float64 (seconds) or string.
func extractDurationField(data map[string]interface{}, key string) string {
	raw, ok := data[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case float64:
		secs := int(v)
		if secs <= 0 {
			return "system default"
		}
		return formatSecondsDuration(secs)
	case string:
		return v
	}
	return fmt.Sprintf("%v", raw)
}

func formatSecondsDuration(secs int) string {
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	if secs < 3600 {
		return fmt.Sprintf("%dm", secs/60)
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
