package vault

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
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

// LogEntry represents a single line from the Vault server log stream.
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Raw       string
}

// MonitorLogs connects to the sys/monitor endpoint and streams log entries.
// The returned channel emits parsed log entries until the context is cancelled.
func (c *Client) MonitorLogs(ctx context.Context, level string) (<-chan LogEntry, error) {
	if level == "" {
		level = "info"
	}

	addr := c.raw.Address()
	url := addr + "/v1/sys/monitor?log_level=" + level
	if fmt := ""; fmt != "" {
		url += "&log_format=" + fmt
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating monitor request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.raw.Token())
	if ns := c.raw.Headers().Get("X-Vault-Namespace"); ns != "" {
		req.Header.Set("X-Vault-Namespace", ns)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to monitor: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("monitor returned status %d", resp.StatusCode)
	}

	ch := make(chan LogEntry, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			entry := parseLogLine(line)
			select {
			case ch <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// parseLogLine extracts timestamp, level, and message from a Vault log line.
// Vault dev log format: "2025-01-15T10:30:00.000Z [INFO]  core: message here"
func parseLogLine(line string) LogEntry {
	entry := LogEntry{Raw: line}

	bracketStart := strings.Index(line, "[")
	bracketEnd := strings.Index(line, "]")
	if bracketStart >= 0 && bracketEnd > bracketStart {
		if bracketStart > 0 {
			ts := strings.TrimSpace(line[:bracketStart])
			entry.Timestamp, _ = time.Parse("2006-01-02T15:04:05.000Z", ts)
			if entry.Timestamp.IsZero() {
				entry.Timestamp, _ = time.Parse(time.RFC3339, ts)
			}
		}
		entry.Level = strings.TrimSpace(line[bracketStart+1 : bracketEnd])
		if bracketEnd+1 < len(line) {
			entry.Message = strings.TrimSpace(line[bracketEnd+1:])
		}
	} else {
		entry.Message = line
	}

	return entry
}
