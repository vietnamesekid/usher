package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/keychain"
	"github.com/vietnamesekid/usher/internal/registry"
	"github.com/vietnamesekid/usher/internal/ui"
	"github.com/vietnamesekid/usher/internal/writers"
)

// Syncer orchestrates the full sync pipeline:
// merge → resolve secrets → backup → write tool configs → inject skills.
type Syncer struct {
	global   config.Config
	project  config.ProjectConfig
	reg      registry.Registry
	kc       keychain.Keychain
	writers  []writers.Writer
	injector *writers.SkillInjector
	promptFn func(string) string
	out      *ui.Output
}

func New(
	global config.Config,
	project config.ProjectConfig,
	reg registry.Registry,
	kc keychain.Keychain,
	promptFn func(string) string,
	out *ui.Output,
) *Syncer {
	return &Syncer{
		global:   global,
		project:  project,
		reg:      reg,
		kc:       kc,
		writers:  writers.All(),
		injector: writers.NewSkillInjector(),
		promptFn: promptFn,
		out:      out,
	}
}

// Run executes the full sync pipeline.
func (s *Syncer) Run() error {
	s.out.Step("Merging config...")
	mc, err := Merge(s.global, s.project, s.reg)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}

	s.out.Step("Resolving secrets...")
	secrets, err := keychain.Resolve(s.global, s.kc, s.promptFn)
	if err != nil {
		return fmt.Errorf("resolve secrets: %w", err)
	}

	rc := buildResolvedConfig(mc, secrets)

	backupsDir := expandHome(mc.BackupsDir)

	var updated []string
	for _, w := range s.writers {
		if !isToolEnabled(w.Name(), rc.Tools) {
			continue
		}
		if !w.Detect() {
			s.out.Info(fmt.Sprintf("skipping %s (binary not found)", w.Name()))
			continue
		}
		if err := w.Backup(backupsDir); err != nil {
			s.out.Warning(fmt.Sprintf("backup %s: %v", w.Name(), err))
		}
		if err := w.Write(rc); err != nil {
			return fmt.Errorf("write %s: %w", w.Name(), err)
		}
		updated = append(updated, w.Name())
	}

	if len(rc.Skills) > 0 {
		s.out.Step("Injecting skills...")
		if err := s.injector.InjectAll(rc); err != nil {
			return fmt.Errorf("inject skills: %w", err)
		}
	}

	s.out.Success(fmt.Sprintf(
		"Synced: %d tool(s) updated (%s), %d MCP(s), %d skill(s)",
		len(updated), strings.Join(updated, ", "),
		len(rc.MCPInstances), len(rc.Skills),
	))
	return nil
}

func buildResolvedConfig(mc MergedConfig, secrets *keychain.ResolvedSecrets) ResolvedConfig {
	rc := ResolvedConfig{
		Tools:      mc.Tools,
		BackupsDir: mc.BackupsDir,
	}
	for _, inst := range mc.Instances {
		token, _ := secrets.Get(inst.AuthKey)
		rc.MCPInstances = append(rc.MCPInstances, ResolvedMCPInstance{
			InstanceName: inst.InstanceName,
			Command:      inst.Command,
			Args:         inst.Args,
			EnvVar:       inst.EnvVar,
			Token:        token,
		})
	}
	for _, sk := range mc.Skills {
		rc.Skills = append(rc.Skills, ResolvedSkill{
			Name:    sk.Name,
			Version: sk.Version,
			Source:  sk.Source,
		})
	}
	return rc
}

func isToolEnabled(name string, tools config.ToolsConfig) bool {
	switch name {
	case "claude":
		return tools.Claude
	case "gemini":
		return tools.Gemini
	case "codex":
		return tools.Codex
	case "cursor":
		return tools.Cursor
	case "windsurf":
		return tools.Windsurf
	case "cline":
		return tools.Cline
	}
	return false
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
