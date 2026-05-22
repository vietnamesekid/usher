package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_GlobalOnly(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.json")
	writeJSON(t, globalPath, Config{
		Version: "1",
		Tools:   ToolsConfig{Claude: true},
		MCPServers: map[string]MCPEntry{
			"supabase": {Instances: []MCPInstance{{Name: "supabase", Enabled: true}}},
		},
	})

	l := &Loader{globalPath: globalPath, projectPath: filepath.Join(dir, "project.json")}
	cfg, hasProject, err := l.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if hasProject {
		t.Error("expected hasProject=false")
	}
	if !cfg.Tools.Claude {
		t.Error("expected Tools.Claude=true")
	}
	if _, ok := cfg.MCPServers["supabase"]; !ok {
		t.Error("expected supabase in MCPServers")
	}
}

func TestLoad_MissingGlobal_ReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	l := &Loader{
		globalPath:  filepath.Join(dir, "config.json"),
		projectPath: filepath.Join(dir, "project.json"),
	}
	cfg, _, err := l.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("default version = %q, want 1", cfg.Version)
	}
}

func TestMergeConfigs_ProjectAddsMCPs(t *testing.T) {
	global := Config{
		Version: "1",
		MCPServers: map[string]MCPEntry{
			"github": {Instances: []MCPInstance{{Name: "github", Enabled: true}}},
		},
		Skills: map[string]SkillEntry{},
	}
	project := ProjectConfig{
		MCPServers: map[string]MCPEntry{
			"supabase": {Instances: []MCPInstance{{Name: "supabase", Enabled: true}}},
		},
	}
	merged := mergeConfigs(global, project)
	if _, ok := merged.MCPServers["github"]; !ok {
		t.Error("merged should contain global MCP 'github'")
	}
	if _, ok := merged.MCPServers["supabase"]; !ok {
		t.Error("merged should contain project MCP 'supabase'")
	}
}

func TestMergeConfigs_ProjectDoesNotOverrideGlobalMCP(t *testing.T) {
	globalInst := MCPInstance{Name: "github", Auth: AuthRef{Key: "global-key"}, Enabled: true}
	projInst := MCPInstance{Name: "github", Auth: AuthRef{Key: "project-key"}, Enabled: true}

	global := Config{
		Version:    "1",
		MCPServers: map[string]MCPEntry{"github": {Instances: []MCPInstance{globalInst}}},
		Skills:     map[string]SkillEntry{},
	}
	project := ProjectConfig{
		MCPServers: map[string]MCPEntry{"github": {Instances: []MCPInstance{projInst}}},
	}
	merged := mergeConfigs(global, project)
	entry := merged.MCPServers["github"]
	if entry.Instances[0].Auth.Key != "global-key" {
		t.Error("project config should not override existing global MCP entry")
	}
}

func TestMergeConfigs_ProjectDisablesSkill(t *testing.T) {
	global := Config{
		Version: "1",
		Skills: map[string]SkillEntry{
			"react":    {Version: "1.0.0"},
			"supabase": {Version: "1.0.0"},
		},
		MCPServers: map[string]MCPEntry{},
	}
	project := ProjectConfig{
		Skills: map[string]SkillEntry{
			"react": {Disabled: true},
		},
	}
	merged := mergeConfigs(global, project)
	if _, ok := merged.Skills["react"]; ok {
		t.Error("disabled skill 'react' should be removed from merged config")
	}
	if _, ok := merged.Skills["supabase"]; !ok {
		t.Error("non-disabled skill 'supabase' should still be present")
	}
}
