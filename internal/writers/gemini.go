package writers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vietnamesekid/usher/internal/types"
)

type geminiMCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// GeminiWriter writes MCP config to ~/.gemini/settings.json.
type GeminiWriter struct{}

func NewGeminiWriter() *GeminiWriter { return &GeminiWriter{} }

func (w *GeminiWriter) Name() string { return "gemini" }

func (w *GeminiWriter) Detect() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

func (w *GeminiWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gemini", "settings.json")
}

func (w *GeminiWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

func (w *GeminiWriter) Write(rc types.ResolvedConfig) error {
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	mcpServers := make(map[string]geminiMCPServer)
	for _, inst := range rc.MCPInstances {
		srv := geminiMCPServer{
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
