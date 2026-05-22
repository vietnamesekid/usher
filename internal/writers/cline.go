package writers

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vietnamesekid/usher/internal/types"
)

// ClineWriter handles skill injection for Cline (VS Code extension).
// Cline has no MCP config format managed by Usher; only skills are supported.
type ClineWriter struct{}

func NewClineWriter() *ClineWriter { return &ClineWriter{} }

func (w *ClineWriter) Name() string { return "cline" }

func (w *ClineWriter) Detect() bool {
	// Detect via the VS Code CLI since Cline is a VS Code extension.
	_, err := exec.LookPath("code")
	return err == nil
}

func (w *ClineWriter) ConfigPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".clinerules")
}

func (w *ClineWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

// Write is a no-op for Cline: MCP config is not managed by Usher.
// Skill injection is handled separately by SkillInjector.
func (w *ClineWriter) Write(_ types.ResolvedConfig) error {
	return nil
}
