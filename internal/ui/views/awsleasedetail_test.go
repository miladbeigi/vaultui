package views

import (
	"strings"
	"testing"
	"time"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestAWSLeaseDetailView_Title(t *testing.T) {
	lease := vault.AWSLease{LeaseID: "aws/creds/deploy-iam-user/abc123"}
	v := NewAWSLeaseDetailView(lease)
	if v.Title() != "Lease: deploy-iam-user" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestAWSLeaseDetailView_View(t *testing.T) {
	lease := vault.AWSLease{
		LeaseID:    "aws/creds/deploy-iam-user/abc123",
		TTL:        30 * time.Minute,
		IssueTime:  "2026-03-15T00:51:10.123456789Z",
		ExpireTime: "2026-03-15T01:21:10.123456789Z",
		Renewable:  true,
	}
	v := NewAWSLeaseDetailView(lease)
	view := v.View(120, 20)

	if !strings.Contains(view, "Lease: deploy-iam-user") {
		t.Error("expected title in view")
	}
}

func TestAWSLeaseDetailView_KeyHints(t *testing.T) {
	v := NewAWSLeaseDetailView(vault.AWSLease{LeaseID: "aws/creds/x/y"})
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected non-empty hints")
	}
}

func TestSplitLeaseID(t *testing.T) {
	tests := []struct {
		id        string
		wantRole  string
		wantShort string
	}{
		{"aws/creds/deploy-iam-user/abc123", "deploy-iam-user", "abc123"},
		{"aws/creds/my-role/xyz", "my-role", "xyz"},
		{"short", "-", "short"},
	}
	for _, tt := range tests {
		role, short := splitLeaseID(tt.id)
		if role != tt.wantRole || short != tt.wantShort {
			t.Errorf("splitLeaseID(%q) = (%q, %q), want (%q, %q)", tt.id, role, short, tt.wantRole, tt.wantShort)
		}
	}
}

func TestFormatLeaseTTL(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "expired"},
		{-1 * time.Second, "expired"},
		{30 * time.Minute, "30m 0s"},
		{2 * time.Hour, "2h 0m"},
	}
	for _, tt := range tests {
		got := formatLeaseTTL(tt.d)
		if got != tt.want {
			t.Errorf("formatLeaseTTL(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
