package writers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vietnamesekid/usher/internal/types"
)

type clineMCPServer struct {
	Command  string            `json:"command,omitempty"`
	Args     []string          `json:"args,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Disabled bool              `json:"disabled,omitempty"`
}

// ClineWriter writes MCP config to the Cline VS Code extension's settings file
// and injects skills into .clinerules via SkillInjector.
type ClineWriter struct{}

func NewClineWriter() *ClineWriter { return &ClineWriter{} }

func (w *ClineWriter) Name() string { return "cline" }

func (w *ClineWriter) Detect() bool {
	_, err := exec.LookPath("code")
	return err == nil
}

// ConfigPath returns the platform-specific path to Cline's MCP settings file.
func (w *ClineWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", "Code", "User",
			"globalStorage", "saoudrizwan.claude-dev", "settings")
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		base = filepath.Join(appdata, "Code", "User",
			"globalStorage", "saoudrizwan.claude-dev", "settings")
	default: // linux
		base = filepath.Join(home, ".config", "Code", "User",
			"globalStorage", "saoudrizwan.claude-dev", "settings")
	}
	return filepath.Join(base, "cline_mcp_settings.json")
}

func (w *ClineWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

func (w *ClineWriter) Write(rc types.ResolvedConfig) error {
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	mcpServers := make(map[string]clineMCPServer)
	for _, inst := range rc.MCPInstances {
		srv := clineMCPServer{
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
