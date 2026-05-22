package writers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vietnamesekid/usher/internal/types"
)

type cursorMCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// CursorWriter writes MCP config to ~/.cursor/mcp.json (global).
// Also updates .cursor/mcp.json in the working directory if it exists.
type CursorWriter struct{}

func NewCursorWriter() *CursorWriter { return &CursorWriter{} }

func (w *CursorWriter) Name() string { return "cursor" }

func (w *CursorWriter) Detect() bool {
	_, err := exec.LookPath("cursor")
	return err == nil
}

func (w *CursorWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "mcp.json")
}

func (w *CursorWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

func (w *CursorWriter) Write(rc types.ResolvedConfig) error {
	if err := w.writeToPath(w.ConfigPath(), rc); err != nil {
		return err
	}
	// Also update project-level .cursor/mcp.json if it exists.
	cwd, _ := os.Getwd()
	projectPath := filepath.Join(cwd, ".cursor", "mcp.json")
	if _, err := os.Stat(projectPath); err == nil {
		return w.writeToPath(projectPath, rc)
	}
	return nil
}

func (w *CursorWriter) writeToPath(path string, rc types.ResolvedConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	mcpServers := make(map[string]cursorMCPServer)
	for _, inst := range rc.MCPInstances {
		srv := cursorMCPServer{
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
