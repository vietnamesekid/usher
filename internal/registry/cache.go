package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type cacheFile[T any] struct {
	FetchedAt time.Time `json:"fetchedAt"`
	Entries   []T       `json:"entries"`
}

type Cache struct {
	dir string
	ttl time.Duration
}

func NewCache(dir string, ttl time.Duration) *Cache {
	return &Cache{dir: dir, ttl: ttl}
}

func (c *Cache) GetMCPs() ([]MCPRegistryEntry, bool, error) {
	return readCache[MCPRegistryEntry](filepath.Join(c.dir, "mcp.json"), c.ttl)
}

func (c *Cache) SetMCPs(entries []MCPRegistryEntry) error {
	return writeCache(filepath.Join(c.dir, "mcp.json"), entries)
}

func (c *Cache) GetSkills() ([]SkillRegistryEntry, bool, error) {
	return readCache[SkillRegistryEntry](filepath.Join(c.dir, "skills.json"), c.ttl)
}

func (c *Cache) SetSkills(entries []SkillRegistryEntry) error {
	return writeCache(filepath.Join(c.dir, "skills.json"), entries)
}

func readCache[T any](path string, ttl time.Duration) ([]T, bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var f cacheFile[T]
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, false, err
	}
	fresh := time.Since(f.FetchedAt) < ttl
	return f.Entries, fresh, nil
}

func writeCache[T any](path string, entries []T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f := cacheFile[T]{FetchedAt: time.Now(), Entries: entries}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
