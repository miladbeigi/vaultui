package vault

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// PathEntry represents a single item in a secret engine path listing.
type PathEntry struct {
	Name  string
	IsDir bool
}

// ListSecrets lists keys under a path in a secret engine.
// For KV v2, it calls LIST /v1/{mount}/metadata/{subPath}.
// For other engines, it calls LIST /v1/{mount}/{subPath}.
// Keys ending with "/" are directories; others are leaf secrets.
func (c *Client) ListSecrets(mount, subPath string, kvV2 bool) ([]PathEntry, error) {
	listPath := mount + subPath
	if kvV2 {
		listPath = mount + "metadata/" + subPath
	}

	cacheKey := "list:" + listPath
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]PathEntry), nil
	}

	secret, err := c.raw.Logical().List(listPath)
	if err != nil {
		return nil, fmt.Errorf("listing path %q: %w", listPath, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"]
	if !ok {
		return nil, nil
	}

	keysList, ok := keysRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys type in response for %q", listPath)
	}

	entries := make([]PathEntry, 0, len(keysList))
	for _, k := range keysList {
		key, ok := k.(string)
		if !ok {
			continue
		}
		entries = append(entries, PathEntry{
			Name:  key,
			IsDir: strings.HasSuffix(key, "/"),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})

	c.cache.Set(cacheKey, entries)
	return entries, nil
}

// VersionEntry represents a single version in KV v2 metadata.
type VersionEntry struct {
	Version      int
	CreatedTime  time.Time
	DeletionTime string
	Destroyed    bool
}

// ReadSecretMetadata reads the metadata for a KV v2 secret, returning version history.
func (c *Client) ReadSecretMetadata(mount, subPath string) ([]VersionEntry, error) {
	metaPath := mount + "metadata/" + subPath
	secret, err := c.raw.Logical().Read(metaPath)
	if err != nil {
		return nil, fmt.Errorf("reading metadata %q: %w", metaPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no metadata found at %q", metaPath)
	}

	versionsRaw, ok := secret.Data["versions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected versions format in metadata for %q", metaPath)
	}

	entries := make([]VersionEntry, 0, len(versionsRaw))
	for vStr, vData := range versionsRaw {
		ver := 0
		fmt.Sscanf(vStr, "%d", &ver)

		vMap, ok := vData.(map[string]interface{})
		if !ok {
			continue
		}

		entry := VersionEntry{Version: ver}

		if ct, ok := vMap["created_time"].(string); ok {
			entry.CreatedTime, _ = time.Parse(time.RFC3339Nano, ct)
		}
		if dt, ok := vMap["deletion_time"].(string); ok && dt != "" {
			entry.DeletionTime = dt
		}
		if d, ok := vMap["destroyed"].(bool); ok {
			entry.Destroyed = d
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Version > entries[j].Version
	})

	return entries, nil
}

// ReadSecretVersion reads a specific version of a KV v2 secret.
func (c *Client) ReadSecretVersion(mount, subPath string, version int) (*SecretData, error) {
	readPath := fmt.Sprintf("%sdata/%s", mount, subPath)
	secret, err := c.raw.Logical().ReadWithData(readPath, map[string][]string{
		"version": {fmt.Sprintf("%d", version)},
	})
	if err != nil {
		return nil, fmt.Errorf("reading secret %q version %d: %w", readPath, version, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secret found at %q version %d", readPath, version)
	}

	raw := secret.Data
	if inner, ok := raw["data"].(map[string]interface{}); ok {
		raw = inner
	}

	sd := &SecretData{Data: make(map[string]string, len(raw))}
	for k, v := range raw {
		sd.Data[k] = fmt.Sprintf("%v", v)
	}
	sd.Keys = make([]string, 0, len(sd.Data))
	for k := range sd.Data {
		sd.Keys = append(sd.Keys, k)
	}
	sort.Strings(sd.Keys)
	return sd, nil
}

// SecretData holds the key-value pairs for a single secret.
type SecretData struct {
	Data map[string]string
	Keys []string
}

// ReadSecret reads a secret at the given path.
// For KV v2 it calls GET /v1/{mount}/data/{subPath},
// for other engines GET /v1/{mount}/{subPath}.
func (c *Client) ReadSecret(mount, subPath string, kvV2 bool) (*SecretData, error) {
	readPath := mount + subPath
	if kvV2 {
		readPath = mount + "data/" + subPath
	}

	secret, err := c.raw.Logical().Read(readPath)
	if err != nil {
		return nil, fmt.Errorf("reading secret %q: %w", readPath, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secret found at %q", readPath)
	}

	raw := secret.Data
	if kvV2 {
		if inner, ok := raw["data"].(map[string]interface{}); ok {
			raw = inner
		}
	}

	sd := &SecretData{Data: make(map[string]string, len(raw))}
	for k, v := range raw {
		sd.Data[k] = fmt.Sprintf("%v", v)
	}

	sd.Keys = make([]string, 0, len(sd.Data))
	for k := range sd.Data {
		sd.Keys = append(sd.Keys, k)
	}
	sort.Strings(sd.Keys)

	return sd, nil
}
