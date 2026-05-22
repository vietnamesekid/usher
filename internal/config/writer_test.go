package config

import (
	"path/filepath"
	"testing"
)

func TestWriter_AddAndRemoveMCPEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	w := NewWriter(path)

	entry := MCPEntry{
		Instances: []MCPInstance{{
			Name:    "supabase",
			Auth:    AuthRef{Type: "keychain", Key: "usher.mcp.supabase.token"},
			Enabled: true,
		}},
	}
	if err := w.AddMCPEntry("supabase", entry); err != nil {
		t.Fatalf("AddMCPEntry: %v", err)
	}

	// Read back and verify.
	l := &Loader{globalPath: path, projectPath: filepath.Join(dir, "project.json")}
	cfg, _, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.MCPServers["supabase"]; !ok {
		t.Fatal("supabase should be present after AddMCPEntry")
	}

	// Remove and verify.
	if err := w.RemoveMCPEntry("supabase"); err != nil {
		t.Fatalf("RemoveMCPEntry: %v", err)
	}
	cfg, _, err = l.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.MCPServers["supabase"]; ok {
		t.Error("supabase should be absent after RemoveMCPEntry")
	}
}

func TestWriter_AddSkillEntry(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(filepath.Join(dir, "config.json"))

	if err := w.AddSkillEntry("supabase", SkillEntry{Version: "1.0.0"}); err != nil {
		t.Fatalf("AddSkillEntry: %v", err)
	}
	l := &Loader{
		globalPath:  filepath.Join(dir, "config.json"),
		projectPath: filepath.Join(dir, "project.json"),
	}
	cfg, _, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Skills["supabase"].Version != "1.0.0" {
		t.Errorf("skill version = %q, want 1.0.0", cfg.Skills["supabase"].Version)
	}
}

func TestWriter_AtomicWrite_PreservesExistingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	w := NewWriter(path)

	// Write initial config with a skill.
	if err := w.AddSkillEntry("react", SkillEntry{Version: "2.0.0"}); err != nil {
		t.Fatal(err)
	}
	// Add MCP — should not lose the skill.
	if err := w.AddMCPEntry("github", MCPEntry{
		Instances: []MCPInstance{{Name: "github", Enabled: true}},
	}); err != nil {
		t.Fatal(err)
	}

	l := &Loader{globalPath: path, projectPath: filepath.Join(dir, "project.json")}
	cfg, _, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Skills["react"]; !ok {
		t.Error("existing skill 'react' was lost after AddMCPEntry")
	}
	if _, ok := cfg.MCPServers["github"]; !ok {
		t.Error("new MCP 'github' should be present")
	}
}
