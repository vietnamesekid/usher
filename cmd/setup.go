package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vietnamesekid/usher/internal/config"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Initialize usher configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
}

var allToolLabels = []string{"Claude Code", "Gemini CLI", "Codex CLI", "Cursor", "Windsurf", "Cline"}

func toolsToSelected(t config.ToolsConfig) []string {
	var s []string
	if t.Claude {
		s = append(s, "Claude Code")
	}
	if t.Gemini {
		s = append(s, "Gemini CLI")
	}
	if t.Codex {
		s = append(s, "Codex CLI")
	}
	if t.Cursor {
		s = append(s, "Cursor")
	}
	if t.Windsurf {
		s = append(s, "Windsurf")
	}
	if t.Cline {
		s = append(s, "Cline")
	}
	return s
}

func selectedToTools(selected []string) config.ToolsConfig {
	set := make(map[string]bool, len(selected))
	for _, s := range selected {
		set[s] = true
	}
	return config.ToolsConfig{
		Claude:   set["Claude Code"],
		Gemini:   set["Gemini CLI"],
		Codex:    set["Codex CLI"],
		Cursor:   set["Cursor"],
		Windsurf: set["Windsurf"],
		Cline:    set["Cline"],
	}
}

func runSetup() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	globalPath := filepath.Join(home, ".usher", "config.json")
	w := config.NewWriter(globalPath)

	_, statErr := os.Stat(globalPath)
	alreadyExists := statErr == nil

	// Load existing tools as default selection when re-running setup.
	var defaults []string
	if alreadyExists {
		loader := config.NewLoader(globalPath)
		if existing, err := loader.LoadGlobal(); err == nil {
			defaults = toolsToSelected(existing.Tools)
		}
	}

	selected := d.prompt.AskMultiSelectWithDefaults(
		"Which AI coding tools do you use? (space to select)",
		allToolLabels,
		defaults,
	)
	tools := selectedToTools(selected)

	if alreadyExists {
		if err := w.SetTools(tools); err != nil {
			return fmt.Errorf("updating config: %w", err)
		}
		d.out.Success("Updated tool selection in " + globalPath)
	} else {
		cfg := config.DefaultConfig()
		cfg.Tools = tools
		if err := w.Init(cfg); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		addGitignoreEntry(".usher/.secrets")
		d.out.Success("Initialized usher at " + globalPath)
		d.out.Info("Next: usher mcp add <name>  or  usher skill add <name>")
	}
	return nil
}

func addGitignoreEntry(entry string) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	gitignorePath := filepath.Join(cwd, ".gitignore")
	data, _ := os.ReadFile(gitignorePath)
	content := string(data)
	for _, line := range splitLines(content) {
		if line == entry {
			return
		}
	}
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	if content != "" && content[len(content)-1] != '\n' {
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, entry)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
