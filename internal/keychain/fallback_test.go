package keychain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFallbackKeychain_SetGetDelete(t *testing.T) {
	kc := NewFallbackKeychain(t.TempDir())

	if err := kc.Set("mykey", "myvalue"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, err := kc.Get("mykey")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "myvalue" {
		t.Errorf("Get = %q, want %q", val, "myvalue")
	}
	ok, err := kc.Exists("mykey")
	if err != nil || !ok {
		t.Errorf("Exists = %v, %v; want true, nil", ok, err)
	}
	if err := kc.Delete("mykey"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, _ = kc.Exists("mykey")
	if ok {
		t.Error("key should not exist after Delete")
	}
}

func TestFallbackKeychain_GetMissing(t *testing.T) {
	kc := NewFallbackKeychain(t.TempDir())
	_, err := kc.Get("nonexistent")
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestFallbackKeychain_MultipleKeys(t *testing.T) {
	kc := NewFallbackKeychain(t.TempDir())
	keys := map[string]string{
		"usher.mcp.supabase.token": "token-abc",
		"usher.mcp.github.token":   "token-xyz",
		"usher.auth.anthropic.key": "key-123",
	}
	for k, v := range keys {
		if err := kc.Set(k, v); err != nil {
			t.Fatalf("Set(%q): %v", k, err)
		}
	}
	for k, want := range keys {
		got, err := kc.Get(k)
		if err != nil {
			t.Fatalf("Get(%q): %v", k, err)
		}
		if got != want {
			t.Errorf("Get(%q) = %q, want %q", k, got, want)
		}
	}
}

func TestFallbackKeychain_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	kc := NewFallbackKeychain(dir)
	if err := kc.Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, ".secrets"))
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf(".secrets permissions = %o, want 0600", perm)
	}
}
