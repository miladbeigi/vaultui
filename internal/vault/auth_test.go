package vault

import (
	"testing"
)

func TestAuthMethod_Constants(t *testing.T) {
	if AuthToken != "token" {
		t.Error("AuthToken should be 'token'")
	}
	if AuthUserpass != "userpass" {
		t.Error("AuthUserpass should be 'userpass'")
	}
	if AuthAppRole != "approle" {
		t.Error("AuthAppRole should be 'approle'")
	}
}

func TestAuthenticate_TokenIsNoop(t *testing.T) {
	c := newTestClient(t)
	err := c.Authenticate(AuthConfig{Method: AuthToken})
	if err != nil {
		t.Errorf("token auth should be no-op, got: %v", err)
	}
}

func TestAuthenticate_EmptyMethodIsNoop(t *testing.T) {
	c := newTestClient(t)
	err := c.Authenticate(AuthConfig{Method: ""})
	if err != nil {
		t.Errorf("empty method should be no-op, got: %v", err)
	}
}

func TestAuthenticate_UnsupportedMethod(t *testing.T) {
	c := newTestClient(t)
	err := c.Authenticate(AuthConfig{Method: "kerberos"})
	if err == nil {
		t.Error("expected error for unsupported method")
	}
}

func TestAuthenticate_UserpassMissingUsername(t *testing.T) {
	c := newTestClient(t)
	err := c.Authenticate(AuthConfig{
		Method:   AuthUserpass,
		Password: "pass",
	})
	if err == nil {
		t.Error("expected error for missing username")
	}
}

func TestAuthenticate_UserpassMissingPassword(t *testing.T) {
	c := newTestClient(t)
	err := c.Authenticate(AuthConfig{
		Method:   AuthUserpass,
		Username: "user",
	})
	if err == nil {
		t.Error("expected error for missing password")
	}
}

func TestAuthenticate_AppRoleMissingRoleID(t *testing.T) {
	c := newTestClient(t)
	err := c.Authenticate(AuthConfig{
		Method: AuthAppRole,
	})
	if err == nil {
		t.Error("expected error for missing role ID")
	}
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClient(ClientConfig{
		Address: "http://127.0.0.1:8200",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}
