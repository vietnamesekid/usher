package writers

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vietnamesekid/usher/internal/types"
)

// WindsurfWriter handles skill injection for Windsurf.
// Windsurf has no MCP config format; only skills are supported.
type WindsurfWriter struct{}

func NewWindsurfWriter() *WindsurfWriter { return &WindsurfWriter{} }

func (w *WindsurfWriter) Name() string { return "windsurf" }

func (w *WindsurfWriter) Detect() bool {
	_, err := exec.LookPath("windsurf")
	return err == nil
}

func (w *WindsurfWriter) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codeium", "windsurf", "memories", "global_rules.md")
}

func (w *WindsurfWriter) Backup(backupsDir string) error {
	return backupFile(w.ConfigPath(), backupsDir, w.Name())
}

// Write is a no-op for Windsurf: MCP is not supported.
// Skill injection is handled separately by SkillInjector.
func (w *WindsurfWriter) Write(_ types.ResolvedConfig) error {
	return nil
}
