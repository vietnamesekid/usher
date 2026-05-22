package config

import (
	"path/filepath"
	"testing"
	"time"
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

	if err := w.AddSkillEntry("supabase", SkillEntry{Source: "supabase/agent-skills"}); err != nil {
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
	if cfg.Skills["supabase"].Source != "supabase/agent-skills" {
		t.Errorf("skill source = %q, want supabase/agent-skills", cfg.Skills["supabase"].Source)
	}
}

func TestWriter_RemoveSkillEntry(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(filepath.Join(dir, "config.json"))

	if err := w.AddSkillEntry("supabase", SkillEntry{Source: "supabase/agent-skills"}); err != nil {
		t.Fatal(err)
	}
	if err := w.RemoveSkillEntry("supabase"); err != nil {
		t.Fatalf("RemoveSkillEntry: %v", err)
	}
	cfg, _ := w.read()
	if _, ok := cfg.Skills["supabase"]; ok {
		t.Error("supabase skill still present after remove")
	}
}

func TestWriter_RemoveAuthEntry(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(filepath.Join(dir, "config.json"))

	for _, e := range []AuthEntry{
		{Provider: "anthropic", KeyRef: "usher.auth.anthropic.api_key", AddedAt: time.Now()},
		{Provider: "google", KeyRef: "usher.auth.google.api_key", AddedAt: time.Now()},
	} {
		if err := w.AddAuthEntry(e); err != nil {
			t.Fatal(err)
		}
	}

	if err := w.RemoveAuthEntry("anthropic"); err != nil {
		t.Fatalf("RemoveAuthEntry: %v", err)
	}

	cfg, _ := w.read()
	for _, a := range cfg.Auth {
		if a.Provider == "anthropic" {
			t.Error("anthropic auth entry still present after revoke")
		}
	}
	found := false
	for _, a := range cfg.Auth {
		if a.Provider == "google" {
			found = true
		}
	}
	if !found {
		t.Error("google auth entry was incorrectly removed")
	}
}

func TestWriter_RemoveAuthEntry_NonExistent_NoError(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(filepath.Join(dir, "config.json"))
	if err := w.RemoveAuthEntry("nonexistent"); err != nil {
		t.Errorf("RemoveAuthEntry on nonexistent provider: %v", err)
	}
}

func TestWriter_SetTools(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(filepath.Join(dir, "config.json"))
	if err := w.Init(DefaultConfig()); err != nil {
		t.Fatal(err)
	}

	tools := ToolsConfig{Claude: true, Cursor: true, Windsurf: true}
	if err := w.SetTools(tools); err != nil {
		t.Fatalf("SetTools: %v", err)
	}

	cfg, _ := w.read()
	if !cfg.Tools.Claude || !cfg.Tools.Cursor || !cfg.Tools.Windsurf {
		t.Error("tools not saved correctly")
	}
	if cfg.Tools.Gemini || cfg.Tools.Codex || cfg.Tools.Cline {
		t.Error("unset tools should be false")
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
