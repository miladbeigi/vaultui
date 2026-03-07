package vault

import (
	"fmt"
	"sort"
)

// TransitKey represents a key in the Transit engine.
type TransitKey struct {
	Name string
}

// TransitKeyDetail holds the details of a transit key.
type TransitKeyDetail struct {
	Name              string
	Type              string
	LatestVersion     int
	MinDecryptVersion int
	MinEncryptVersion int
	Exportable        bool
	DeletionAllowed   bool
}

// ListTransitKeys lists keys from a Transit engine mount.
func (c *Client) ListTransitKeys(mount string) ([]TransitKey, error) {
	cacheKey := "transit:keys:" + mount
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]TransitKey), nil
	}

	secret, err := c.raw.Logical().List(mount + "keys")
	if err != nil {
		return nil, fmt.Errorf("listing transit keys at %q: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	keys := make([]TransitKey, 0, len(keysRaw))
	for _, k := range keysRaw {
		if name, ok := k.(string); ok {
			keys = append(keys, TransitKey{Name: name})
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Name < keys[j].Name
	})

	c.cache.Set(cacheKey, keys)
	return keys, nil
}

// ReadTransitKey reads the details of a specific transit key.
func (c *Client) ReadTransitKey(mount, name string) (*TransitKeyDetail, error) {
	readPath := mount + "keys/" + name
	secret, err := c.raw.Logical().Read(readPath)
	if err != nil {
		return nil, fmt.Errorf("reading transit key %q: %w", readPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no transit key found at %q", readPath)
	}

	detail := &TransitKeyDetail{Name: name}
	if v, ok := secret.Data["type"].(string); ok {
		detail.Type = v
	}
	if v, ok := secret.Data["latest_version"].(float64); ok {
		detail.LatestVersion = int(v)
	}
	if v, ok := secret.Data["min_decryption_version"].(float64); ok {
		detail.MinDecryptVersion = int(v)
	}
	if v, ok := secret.Data["min_encryption_version"].(float64); ok {
		detail.MinEncryptVersion = int(v)
	}
	if v, ok := secret.Data["exportable"].(bool); ok {
		detail.Exportable = v
	}
	if v, ok := secret.Data["deletion_allowed"].(bool); ok {
		detail.DeletionAllowed = v
	}

	return detail, nil
}

// TransitEncrypt encrypts plaintext using a transit key.
func (c *Client) TransitEncrypt(mount, keyName, plaintext string) (string, error) {
	path := mount + "encrypt/" + keyName
	secret, err := c.raw.Logical().Write(path, map[string]interface{}{
		"plaintext": plaintext,
	})
	if err != nil {
		return "", fmt.Errorf("transit encrypt: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("transit encrypt returned no data")
	}
	ciphertext, ok := secret.Data["ciphertext"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected ciphertext format")
	}
	return ciphertext, nil
}

// TransitDecrypt decrypts ciphertext using a transit key.
func (c *Client) TransitDecrypt(mount, keyName, ciphertext string) (string, error) {
	path := mount + "decrypt/" + keyName
	secret, err := c.raw.Logical().Write(path, map[string]interface{}{
		"ciphertext": ciphertext,
	})
	if err != nil {
		return "", fmt.Errorf("transit decrypt: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("transit decrypt returned no data")
	}
	plaintext, ok := secret.Data["plaintext"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected plaintext format")
	}
	return plaintext, nil
}
