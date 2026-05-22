package writers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vietnamesekid/usher/internal/types"
)

type windsurfMCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// WindsurfWriter writes MCP config to ~/.codeium/windsurf/mcp_config.json
// and injects skills into the global rules file via SkillInjector.
type WindsurfWriter struct{}

func NewWindsurfWriter() *WindsurfWriter { return &WindsurfWriter{} }

func (w *WindsurfWriter) Name() string { return "windsurf" }

func (w *WindsurfWriter) Detect() bool {
	// Windsurf ships a "windsurf" CLI on macOS/Linux; on Windows it may not be in PATH.
	if runtime.GOOS == "windows" {
		// Fall back to checking the config directory existence.
		home, _ := os.UserHomeDir()
		_, err := os.Stat(filepath.Join(home, ".codeium", "windsurf"))
		return err == nil
	}
	_, err := exec.LookPath("windsurf")
	return err == nil
}

// ConfigPath returns the MCP config file path.
func (w *WindsurfWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
}

func (w *WindsurfWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

func (w *WindsurfWriter) Write(rc types.ResolvedConfig) error {
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	mcpServers := make(map[string]windsurfMCPServer)
	for _, inst := range rc.MCPInstances {
		srv := windsurfMCPServer{
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
