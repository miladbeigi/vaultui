package vault

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("key1", "value1")

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v != "value1" {
		t.Errorf("expected 'value1', got %v", v)
	}
}

func TestCache_Miss(t *testing.T) {
	c := NewCache(1 * time.Minute)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestCache_Expiry(t *testing.T) {
	c := NewCache(1 * time.Millisecond)
	c.Set("key1", "value1")

	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	c.Invalidate("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be invalidated")
	}

	_, ok = c.Get("key2")
	if !ok {
		t.Error("expected key2 to still be cached")
	}
}

func TestCache_InvalidatePrefix(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("list:secret/apps/", "v1")
	c.Set("list:secret/infra/", "v2")
	c.Set("sys/health", "v3")

	c.InvalidatePrefix("list:secret/")

	if _, ok := c.Get("list:secret/apps/"); ok {
		t.Error("expected apps cache to be invalidated")
	}
	if _, ok := c.Get("list:secret/infra/"); ok {
		t.Error("expected infra cache to be invalidated")
	}
	if _, ok := c.Get("sys/health"); !ok {
		t.Error("expected sys/health to still be cached")
	}
}

func TestCache_Clear(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}
}

func TestCache_Size(t *testing.T) {
	c := NewCache(1 * time.Minute)
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}

	c.Set("a", 1)
	c.Set("b", 2)
	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("key", "old")
	c.Set("key", "new")

	v, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v != "new" {
		t.Errorf("expected 'new', got %v", v)
	}
}
