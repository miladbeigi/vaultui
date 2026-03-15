package vault

import (
	"testing"
)

func TestFormatSecondsDuration(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{0, "0s"},
		{30, "30s"},
		{60, "1m"},
		{90, "1m"},
		{3600, "1h"},
		{3660, "1h 1m"},
		{86400, "24h"},
		{90000, "25h"},
	}
	for _, tt := range tests {
		got := formatSecondsDuration(tt.secs)
		if got != tt.want {
			t.Errorf("formatSecondsDuration(%d) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

func TestExtractStringSlice(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		key  string
		want int
	}{
		{
			name: "nil data",
			data: map[string]interface{}{},
			key:  "roles",
			want: 0,
		},
		{
			name: "interface slice",
			data: map[string]interface{}{
				"roles": []interface{}{"a", "b", "c"},
			},
			key:  "roles",
			want: 3,
		},
		{
			name: "comma-separated string",
			data: map[string]interface{}{
				"roles": "a,b",
			},
			key:  "roles",
			want: 2,
		},
		{
			name: "empty string",
			data: map[string]interface{}{
				"roles": "",
			},
			key:  "roles",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractStringSlice(tt.data, tt.key)
			if len(got) != tt.want {
				t.Errorf("extractStringSlice() returned %d items, want %d", len(got), tt.want)
			}
		})
	}
}

func TestExtractDurationField(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		key  string
		want string
	}{
		{
			name: "missing key",
			data: map[string]interface{}{},
			key:  "ttl",
			want: "",
		},
		{
			name: "zero float",
			data: map[string]interface{}{"ttl": float64(0)},
			key:  "ttl",
			want: "system default",
		},
		{
			name: "3600 seconds",
			data: map[string]interface{}{"ttl": float64(3600)},
			key:  "ttl",
			want: "1h",
		},
		{
			name: "string value",
			data: map[string]interface{}{"ttl": "1h30m"},
			key:  "ttl",
			want: "1h30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDurationField(tt.data, tt.key)
			if got != tt.want {
				t.Errorf("extractDurationField() = %q, want %q", got, tt.want)
			}
		})
	}
}
