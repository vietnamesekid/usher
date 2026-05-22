package config

import "time"

type AuthRef struct {
	Type string `json:"type"` // "keychain" | "env"
	Key  string `json:"key"`
}

type MCPInstance struct {
	Name    string  `json:"name"`
	Auth    AuthRef `json:"auth"`
	Enabled bool    `json:"enabled"`
}

type MCPEntry struct {
	Instances []MCPInstance `json:"instances"`
}

type SkillEntry struct {
	Version  string `json:"version"`
	Disabled bool   `json:"disabled,omitempty"`
}

type AuthEntry struct {
	Provider string    `json:"provider"`
	KeyRef   string    `json:"keyRef"`
	AddedAt  time.Time `json:"addedAt"`
}

type ToolsConfig struct {
	Claude bool `json:"claude"`
	Gemini bool `json:"gemini"`
	Codex  bool `json:"codex"`
	Cursor bool `json:"cursor"`
}

type SyncConfig struct {
	BackupsDir string `json:"backupsDir,omitempty"`
}

type Config struct {
	Version    string               `json:"version"`
	Tools      ToolsConfig          `json:"tools"`
	MCPServers map[string]MCPEntry  `json:"mcpServers,omitempty"`
	Skills     map[string]SkillEntry `json:"skills,omitempty"`
	Auth       []AuthEntry          `json:"auth,omitempty"`
	Sync       SyncConfig           `json:"sync,omitempty"`
}

type ProjectConfig struct {
	MCPServers map[string]MCPEntry  `json:"mcpServers,omitempty"`
	Skills     map[string]SkillEntry `json:"skills,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Version: "1",
		Tools: ToolsConfig{
			Claude: true,
		},
		MCPServers: make(map[string]MCPEntry),
		Skills:     make(map[string]SkillEntry),
		Auth:       []AuthEntry{},
	}
}
