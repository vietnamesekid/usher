package keychain

import (
	"fmt"
	"os"
)

// Keychain is the interface for storing and retrieving secrets.
type Keychain interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	Exists(key string) (bool, error)
}

const ServiceName = "usher"

// MCPTokenKey returns the canonical keychain key for an MCP instance token.
func MCPTokenKey(instanceName string) string {
	return fmt.Sprintf("usher.mcp.%s.token", instanceName)
}

// AuthKey returns the canonical keychain key for a provider credential.
func AuthKey(provider, keyType string) string {
	return fmt.Sprintf("usher.auth.%s.%s", provider, keyType)
}

// New returns the appropriate Keychain implementation.
//
// Selection order:
//  1. USHER_KEYCHAIN_FALLBACK=1 env var → FallbackKeychain
//  2. System keychain (go-keyring) — probe with a no-op call; if it fails → FallbackKeychain
//  3. Otherwise → SystemKeychain
func New(fallbackDir string) Keychain {
	if os.Getenv("USHER_KEYCHAIN_FALLBACK") == "1" {
		return NewFallbackKeychain(fallbackDir)
	}

	// Probe the system keychain. go-keyring returns an error on Linux without
	// libsecret, or in headless CI environments without a keyring daemon.
	sys := NewSystemKeychain()
	if err := probeSystemKeychain(sys); err != nil {
		return NewFallbackKeychain(fallbackDir)
	}
	return sys
}

// probeSystemKeychain attempts a benign read to verify the system keychain is
// available. ErrNotFound is expected and OK — any other error means unusable.
func probeSystemKeychain(kc *SystemKeychain) error {
	_, err := kc.Get("__usher_probe__")
	if err == nil {
		return nil
	}
	// ErrNotFound means keychain is reachable but key doesn't exist — that's fine.
	if isNotFound(err) {
		return nil
	}
	return err
}
