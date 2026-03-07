package vault

import (
	"testing"
)

func TestStartTokenRenewer_StopsCleanly(t *testing.T) {
	c := newTestClient(t)
	r := StartTokenRenewer(c)
	if r == nil {
		t.Fatal("expected non-nil renewer")
	}
	r.Stop()
}

func TestStartTokenRenewer_DoubleStop(t *testing.T) {
	c := newTestClient(t)
	r := StartTokenRenewer(c)
	r.Stop()
	r.Stop()
}
