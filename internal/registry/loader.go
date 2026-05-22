package registry

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

// NotFoundError is returned when a registry entry is not found,
// with suggestions for similar names.
type NotFoundError struct {
	Name        string
	Suggestions []string
}

func (e *NotFoundError) Error() string {
	msg := fmt.Sprintf("MCP %q not found in registry", e.Name)
	if len(e.Suggestions) > 0 {
		msg += fmt.Sprintf(". Did you mean: %s?", strings.Join(e.Suggestions, ", "))
	}
	return msg
}

type SkillNotFoundError struct {
	Name        string
	Suggestions []string
}

func (e *SkillNotFoundError) Error() string {
	msg := fmt.Sprintf("skill %q not found in registry", e.Name)
	if len(e.Suggestions) > 0 {
		msg += fmt.Sprintf(". Did you mean: %s?", strings.Join(e.Suggestions, ", "))
	}
	return msg
}

// Loader implements Registry using embedded JSON files + optional cache.
type Loader struct {
	cache *Cache
}

func New() *Loader {
	return &Loader{}
}

func NewWithCache(c *Cache) *Loader {
	return &Loader{cache: c}
}

func (l *Loader) GetMCP(name string) (MCPRegistryEntry, error) {
	entries := l.ListMCPs()
	for _, e := range entries {
		if e.Name == name {
			return e, nil
		}
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return MCPRegistryEntry{}, &NotFoundError{
		Name:        name,
		Suggestions: suggestSimilar(name, names),
	}
}

func (l *Loader) GetSkill(name string) (SkillRegistryEntry, error) {
	entries := l.ListSkills()
	for _, e := range entries {
		if e.Name == name {
			return e, nil
		}
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return SkillRegistryEntry{}, &SkillNotFoundError{
		Name:        name,
		Suggestions: suggestSimilar(name, names),
	}
}

func (l *Loader) ListMCPs() []MCPRegistryEntry {
	if l.cache != nil {
		if entries, fresh, _ := l.cache.GetMCPs(); fresh {
			return entries
		}
	}
	entries, _ := loadBundledMCPs()
	if l.cache != nil && len(entries) > 0 {
		_ = l.cache.SetMCPs(entries)
	}
	return entries
}

func (l *Loader) ListSkills() []SkillRegistryEntry {
	if l.cache != nil {
		if entries, fresh, _ := l.cache.GetSkills(); fresh {
			return entries
		}
	}
	entries, _ := loadBundledSkills()
	if l.cache != nil && len(entries) > 0 {
		_ = l.cache.SetSkills(entries)
	}
	return entries
}

func loadBundledMCPs() ([]MCPRegistryEntry, error) {
	var entries []MCPRegistryEntry
	err := fs.WalkDir(mcpFS, "mcp", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		data, err := mcpFS.ReadFile(path)
		if err != nil {
			return err
		}
		var entry MCPRegistryEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		entries = append(entries, entry)
		return nil
	})
	return entries, err
}

func loadBundledSkills() ([]SkillRegistryEntry, error) {
	var entries []SkillRegistryEntry
	err := fs.WalkDir(skillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		data, err := skillsFS.ReadFile(path)
		if err != nil {
			return err
		}
		var entry SkillRegistryEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		entries = append(entries, entry)
		return nil
	})
	return entries, err
}

// suggestSimilar returns names within Levenshtein distance 2 of query.
func suggestSimilar(query string, names []string) []string {
	var suggestions []string
	for _, name := range names {
		if levenshtein(query, name) <= 2 {
			suggestions = append(suggestions, name)
		}
	}
	return suggestions
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	ra, rb := []rune(a), []rune(b)
	m, n := len(ra), len(rb)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		dp[i][0] = i
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if ra[i-1] == rb[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min3(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}
	return dp[m][n]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
