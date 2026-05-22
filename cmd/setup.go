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

func runSetup() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	usherDir := filepath.Join(home, ".usher")

	globalPath := filepath.Join(usherDir, "config.json")
	if _, err := os.Stat(globalPath); err == nil {
		d.out.Info("Usher is already initialized at " + globalPath)
		return nil
	}

	allTools := []string{"Claude Code", "Gemini CLI", "Codex CLI", "Cursor", "Windsurf", "Cline"}
	selected := d.prompt.AskMultiSelect("Which AI coding tools do you use? (space to select)", allTools)

	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}
	tools := config.ToolsConfig{
		Claude:   selectedSet["Claude Code"],
		Gemini:   selectedSet["Gemini CLI"],
		Codex:    selectedSet["Codex CLI"],
		Cursor:   selectedSet["Cursor"],
		Windsurf: selectedSet["Windsurf"],
		Cline:    selectedSet["Cline"],
	}

	cfg := config.DefaultConfig()
	cfg.Tools = tools

	w := config.NewWriter(globalPath)
	if err := w.Init(cfg); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Add .usher/.secrets to .gitignore in cwd.
	addGitignoreEntry(".usher/.secrets")

	d.out.Success("Initialized usher at " + globalPath)
	d.out.Info("Next: usher mcp add <name>  or  usher skill add <name>")
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
