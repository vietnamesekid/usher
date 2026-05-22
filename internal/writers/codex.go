package writers

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vietnamesekid/usher/internal/types"
)

type codexMCPServer struct {
	Name    string            `toml:"name"`
	Command string            `toml:"command"`
	Args    []string          `toml:"args"`
	Env     map[string]string `toml:"env,omitempty"`
}

type codexConfig struct {
	MCPServers []codexMCPServer `toml:"mcp_servers"`
}

// CodexWriter writes MCP config to ~/.codex/config.toml.
type CodexWriter struct{}

func NewCodexWriter() *CodexWriter { return &CodexWriter{} }

func (w *CodexWriter) Name() string { return "codex" }

func (w *CodexWriter) Detect() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

func (w *CodexWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex", "config.toml")
}

func (w *CodexWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

func (w *CodexWriter) Write(rc types.ResolvedConfig) error {
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Read existing TOML into a raw map to preserve non-MCP keys.
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		_ = toml.Unmarshal(data, &existing)
	}

	var servers []codexMCPServer
	for _, inst := range rc.MCPInstances {
		srv := codexMCPServer{
			Name:    inst.InstanceName,
			Command: inst.Command,
			Args:    inst.Args,
		}
		if inst.EnvVar != "" && inst.Token != "" {
			srv.Env = map[string]string{inst.EnvVar: inst.Token}
		}
		servers = append(servers, srv)
	}
	existing["mcp_servers"] = servers

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(existing); err != nil {
		return err
	}
	return atomicWrite(path, buf.Bytes())
}
