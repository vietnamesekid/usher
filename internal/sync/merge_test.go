package sync

import (
	"testing"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/registry"
)

// stubRegistry implements registry.Registry for testing.
type stubRegistry struct {
	mcps map[string]registry.MCPRegistryEntry
}

func (r *stubRegistry) GetMCP(name string) (registry.MCPRegistryEntry, error) {
	e, ok := r.mcps[name]
	if !ok {
		return registry.MCPRegistryEntry{}, &registry.NotFoundError{Name: name}
	}
	return e, nil
}
func (r *stubRegistry) GetSkill(name string) (registry.SkillRegistryEntry, error) {
	return registry.SkillRegistryEntry{}, &registry.SkillNotFoundError{Name: name}
}
func (r *stubRegistry) ListMCPs() []registry.MCPRegistryEntry  { return nil }
func (r *stubRegistry) ListSkills() []registry.SkillRegistryEntry { return nil }

func reg(mcps map[string]registry.MCPRegistryEntry) registry.Registry {
	return &stubRegistry{mcps: mcps}
}

func TestMerge_SkillsFromConfig_NoRegistryLookup(t *testing.T) {
	// Registry has no skills — merge should still succeed because skills
	// no longer go through the registry.
	global := config.Config{
		Version: "1",
		Skills: map[string]config.SkillEntry{
			"supabase": {Source: "supabase/agent-skills"},
			"react":    {Source: "some/react-skills"},
		},
		MCPServers: map[string]config.MCPEntry{},
	}
	r := reg(nil) // empty registry

	mc, err := Merge(global, config.ProjectConfig{}, r)
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	if len(mc.Skills) != 2 {
		t.Fatalf("got %d skills, want 2", len(mc.Skills))
	}

	byName := map[string]MergedSkill{}
	for _, s := range mc.Skills {
		byName[s.Name] = s
	}
	if byName["supabase"].Source != "supabase/agent-skills" {
		t.Errorf("supabase source = %q, want supabase/agent-skills", byName["supabase"].Source)
	}
	if byName["react"].Source != "some/react-skills" {
		t.Errorf("react source = %q, want some/react-skills", byName["react"].Source)
	}
}

func TestMerge_DisabledSkillsExcluded(t *testing.T) {
	global := config.Config{
		Version: "1",
		Skills: map[string]config.SkillEntry{
			"supabase": {Source: "supabase/agent-skills"},
			"disabled": {Source: "x/y", Disabled: true},
		},
		MCPServers: map[string]config.MCPEntry{},
	}

	mc, err := Merge(global, config.ProjectConfig{}, reg(nil))
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}
	for _, s := range mc.Skills {
		if s.Name == "disabled" {
			t.Error("disabled skill should not appear in merged skills")
		}
	}
	if len(mc.Skills) != 1 {
		t.Errorf("got %d skills, want 1", len(mc.Skills))
	}
}

func TestMerge_ProjectDisablesGlobalSkill(t *testing.T) {
	global := config.Config{
		Version: "1",
		Skills: map[string]config.SkillEntry{
			"supabase": {Source: "supabase/agent-skills"},
		},
		MCPServers: map[string]config.MCPEntry{},
	}
	project := config.ProjectConfig{
		Skills: map[string]config.SkillEntry{
			"supabase": {Disabled: true},
		},
	}

	mc, err := Merge(global, project, reg(nil))
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}
	if len(mc.Skills) != 0 {
		t.Errorf("got %d skills, want 0 (supabase disabled by project)", len(mc.Skills))
	}
}

func TestMerge_MCPsExpandedFromRegistry(t *testing.T) {
	global := config.Config{
		Version: "1",
		MCPServers: map[string]config.MCPEntry{
			"github": {Instances: []config.MCPInstance{
				{Name: "github", Auth: config.AuthRef{Type: "keychain", Key: "usher.mcp.github.token"}, Enabled: true},
			}},
		},
		Skills: map[string]config.SkillEntry{},
	}
	r := reg(map[string]registry.MCPRegistryEntry{
		"github": {
			Name:    "github",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-github@latest"},
			Auth:    registry.MCPAuth{EnvVar: "GITHUB_PERSONAL_ACCESS_TOKEN"},
		},
	})

	mc, err := Merge(global, config.ProjectConfig{}, r)
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}
	if len(mc.Instances) != 1 {
		t.Fatalf("got %d instances, want 1", len(mc.Instances))
	}
	inst := mc.Instances[0]
	if inst.Command != "npx" {
		t.Errorf("command = %q, want npx", inst.Command)
	}
	if inst.EnvVar != "GITHUB_PERSONAL_ACCESS_TOKEN" {
		t.Errorf("envVar = %q, want GITHUB_PERSONAL_ACCESS_TOKEN", inst.EnvVar)
	}
	if inst.AuthKey != "usher.mcp.github.token" {
		t.Errorf("authKey = %q, want usher.mcp.github.token", inst.AuthKey)
	}
}
