package vault

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// AuditDevice represents a configured audit backend.
type AuditDevice struct {
	Path        string
	Type        string
	Description string
	Options     map[string]string
}

// ListAuditDevices returns all enabled audit devices, sorted by path.
func (c *Client) ListAuditDevices() ([]AuditDevice, error) {
	if v, ok := c.cache.Get("sys/audit"); ok {
		return v.([]AuditDevice), nil
	}

	secret, err := c.raw.Logical().Read("sys/audit")
	if err != nil {
		return nil, fmt.Errorf("listing audit devices: %w", err)
	}

	var devices []AuditDevice
	if secret != nil && secret.Data != nil {
		for path, raw := range secret.Data {
			m, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			d := AuditDevice{Path: path}
			if v, ok := m["type"].(string); ok {
				d.Type = v
			}
			if v, ok := m["description"].(string); ok {
				d.Description = v
			}
			if opts, ok := m["options"].(map[string]interface{}); ok {
				d.Options = make(map[string]string)
				for k, v := range opts {
					d.Options[k] = fmt.Sprintf("%v", v)
				}
			}
			devices = append(devices, d)
		}
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Path < devices[j].Path
	})

	c.cache.Set("sys/audit", devices)
	return devices, nil
}

// AuditEntry represents a parsed audit log entry from Vault.
type AuditEntry struct {
	Time      time.Time
	Type      string // "request" or "response"
	Operation string
	Path      string
	SourceIP  string
	Error     string
}

// rawAuditEntry is the JSON structure of a Vault audit log line.
type rawAuditEntry struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Request struct {
		Operation     string `json:"operation"`
		Path          string `json:"path"`
		RemoteAddress string `json:"remote_address"`
	} `json:"request"`
	Response struct {
	} `json:"response"`
	Error string `json:"error"`
}

// AuditLogFilePath returns the file path of the first file-type audit device,
// or empty string if none is configured.
func (c *Client) AuditLogFilePath() string {
	devices, err := c.ListAuditDevices()
	if err != nil {
		return ""
	}
	for _, d := range devices {
		if d.Type == "file" && d.Options != nil {
			if fp, ok := d.Options["file_path"]; ok && fp != "stdout" && fp != "stderr" {
				return fp
			}
		}
	}
	return ""
}

// TailAuditLog opens the audit log file and streams new entries as they appear.
// It seeks to the end of the file and only reads new entries.
func TailAuditLog(ctx context.Context, filePath string) (<-chan AuditEntry, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening audit log %q: %w", filePath, err)
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		f.Close()
		return nil, fmt.Errorf("seeking to end of audit log: %w", err)
	}

	ch := make(chan AuditEntry, 64)
	go func() {
		defer f.Close()
		defer close(ch)

		reader := bufio.NewReader(f)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					time.Sleep(200 * time.Millisecond)
					continue
				}
				return
			}

			entry, ok := parseAuditJSON(line)
			if !ok {
				continue
			}

			select {
			case ch <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func parseAuditJSON(data []byte) (AuditEntry, bool) {
	var raw rawAuditEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return AuditEntry{}, false
	}

	entry := AuditEntry{
		Type:      raw.Type,
		Operation: raw.Request.Operation,
		Path:      raw.Request.Path,
		SourceIP:  raw.Request.RemoteAddress,
		Error:     raw.Error,
	}

	entry.Time, _ = time.Parse(time.RFC3339Nano, raw.Time)

	return entry, true
}
