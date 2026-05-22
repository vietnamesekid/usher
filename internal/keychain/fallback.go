package keychain

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FallbackKeychain stores secrets in ~/.usher/.secrets (chmod 600).
// Used when no system keychain is available (headless servers, Docker).
type FallbackKeychain struct {
	path string
}

func NewFallbackKeychain(dir string) *FallbackKeychain {
	return &FallbackKeychain{path: filepath.Join(dir, ".secrets")}
}

func (f *FallbackKeychain) Get(key string) (string, error) {
	m, err := f.readAll()
	if err != nil {
		return "", err
	}
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}
	return v, nil
}

func (f *FallbackKeychain) Set(key, value string) error {
	m, err := f.readAll()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if m == nil {
		m = make(map[string]string)
	}
	m[key] = value
	return f.writeAll(m)
}

func (f *FallbackKeychain) Delete(key string) error {
	m, err := f.readAll()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	delete(m, key)
	return f.writeAll(m)
}

func (f *FallbackKeychain) Exists(key string) (bool, error) {
	m, err := f.readAll()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	_, ok := m[key]
	return ok, nil
}

func (f *FallbackKeychain) readAll() (map[string]string, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	m := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m, scanner.Err()
}

func (f *FallbackKeychain) writeAll(m map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0700); err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("# Usher secrets — do not commit this file\n")
	for k, v := range m {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	return os.WriteFile(f.path, []byte(sb.String()), 0600)
}
