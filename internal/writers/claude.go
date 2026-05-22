package writers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vietnamesekid/usher/internal/types"
)

type claudeMCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ClaudeWriter writes MCP config to ~/.claude/settings.json.
// Unknown fields in the existing file are preserved.
type ClaudeWriter struct{}

func NewClaudeWriter() *ClaudeWriter { return &ClaudeWriter{} }

func (w *ClaudeWriter) Name() string { return "claude" }

func (w *ClaudeWriter) Detect() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (w *ClaudeWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func (w *ClaudeWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

func (w *ClaudeWriter) Write(rc types.ResolvedConfig) error {
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Read existing file into a raw map to preserve unknown fields.
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	// Build mcpServers map.
	mcpServers := make(map[string]claudeMCPServer)
	for _, inst := range rc.MCPInstances {
		srv := claudeMCPServer{
			Command: inst.Command,
			Args:    inst.Args,
		}
		if inst.EnvVar != "" && inst.Token != "" {
			srv.Env = map[string]string{inst.EnvVar: inst.Token}
		}
		mcpServers[inst.InstanceName] = srv
	}

	mcpJSON, err := json.Marshal(mcpServers)
	if err != nil {
		return err
	}
	raw["mcpServers"] = mcpJSON

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}
