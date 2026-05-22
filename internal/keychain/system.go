package keychain

import "github.com/zalando/go-keyring"

func isNotFound(err error) bool {
	return err == keyring.ErrNotFound
}

// SystemKeychain uses the OS-native secret store via go-keyring.
// On macOS: Keychain Access. On Linux: libsecret. On Windows: Credential Manager.
type SystemKeychain struct{}

func NewSystemKeychain() *SystemKeychain {
	return &SystemKeychain{}
}

func (s *SystemKeychain) Get(key string) (string, error) {
	return keyring.Get(ServiceName, key)
}

func (s *SystemKeychain) Set(key, value string) error {
	return keyring.Set(ServiceName, key, value)
}

func (s *SystemKeychain) Delete(key string) error {
	return keyring.Delete(ServiceName, key)
}

func (s *SystemKeychain) Exists(key string) (bool, error) {
	_, err := keyring.Get(ServiceName, key)
	if err == keyring.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
