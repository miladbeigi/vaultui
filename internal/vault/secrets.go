package vault

import (
	"fmt"
	"sort"
	"strings"
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
		// Directories first, then alphabetical
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}
