package keychain

import (
	"fmt"

	"github.com/vietnamesekid/usher/internal/config"
)

// ResolvedSecrets holds plaintext secrets keyed by their keychain key.
// This struct must NEVER be serialized to disk.
type ResolvedSecrets struct {
	m map[string]string
}

func (r *ResolvedSecrets) Get(key string) (string, bool) {
	v, ok := r.m[key]
	return v, ok
}

// Resolve reads all auth references in cfg and fetches their values from kc.
// If a key is missing and promptFn is non-nil, the user is prompted and the
// value is stored back into the keychain. Pass nil promptFn to error on missing.
func Resolve(cfg config.Config, kc Keychain, promptFn func(question string) string) (*ResolvedSecrets, error) {
	rs := &ResolvedSecrets{m: make(map[string]string)}

	for serverName, entry := range cfg.MCPServers {
		for _, inst := range entry.Instances {
			if inst.Auth.Type != "keychain" {
				continue
			}
			key := inst.Auth.Key
			if _, already := rs.m[key]; already {
				continue
			}
			val, err := kc.Get(key)
			if err != nil {
				if promptFn == nil {
					return nil, fmt.Errorf("keychain key %q for %s not found: %w", key, serverName, err)
				}
				val = promptFn(fmt.Sprintf("Enter token for %s (key: %s):", serverName, key))
				if val != "" {
					_ = kc.Set(key, val)
				}
			}
			rs.m[key] = val
		}
	}

	return rs, nil
}
