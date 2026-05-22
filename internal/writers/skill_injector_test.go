package writers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/types"
)

// installSkill writes a SKILL.md into a temp agents/skills/<name>/ dir and
// returns a t.Setenv restorer that points HOME to the temp dir.
func installSkill(t *testing.T, name, content string) {
	t.Helper()
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".agents", "skills", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
}

func TestSkillInjector_InjectAll_AppendsBlock(t *testing.T) {
	installSkill(t, "supabase", "Use Supabase for your database needs.")

	dir := t.TempDir()
	claudeMD := filepath.Join(dir, "CLAUDE.md")

	// Change working directory so instructionFiles() picks up dir.
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	rc := types.ResolvedConfig{
		Tools:  config.ToolsConfig{Claude: true},
		Skills: []types.ResolvedSkill{{Name: "supabase", Source: "supabase/agent-skills"}},
	}

	si := NewSkillInjector()
	if err := si.InjectAll(rc); err != nil {
		t.Fatalf("InjectAll: %v", err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, "<!-- usher:skill:supabase:start -->") {
		t.Error("start marker missing from CLAUDE.md")
	}
	if !strings.Contains(got, "<!-- usher:skill:supabase:end -->") {
		t.Error("end marker missing from CLAUDE.md")
	}
	if !strings.Contains(got, "Use Supabase for your database needs.") {
		t.Error("skill content missing from CLAUDE.md")
	}
}

func TestSkillInjector_InjectAll_UpdatesExistingBlock(t *testing.T) {
	installSkill(t, "supabase", "Updated content.")

	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	claudeMD := filepath.Join(dir, "CLAUDE.md")
	initial := "# My file\n\n<!-- usher:skill:supabase:start -->\nOld content.\n<!-- usher:skill:supabase:end -->\n"
	if err := os.WriteFile(claudeMD, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	rc := types.ResolvedConfig{
		Tools:  config.ToolsConfig{Claude: true},
		Skills: []types.ResolvedSkill{{Name: "supabase"}},
	}
	if err := NewSkillInjector().InjectAll(rc); err != nil {
		t.Fatalf("InjectAll: %v", err)
	}

	data, _ := os.ReadFile(claudeMD)
	got := string(data)

	if strings.Contains(got, "Old content.") {
		t.Error("old content should have been replaced")
	}
	if !strings.Contains(got, "Updated content.") {
		t.Error("updated content missing")
	}
	if !strings.Contains(got, "# My file") {
		t.Error("content before markers was lost")
	}
}

func TestSkillInjector_InjectAll_MissingLocalFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	rc := types.ResolvedConfig{
		Tools:  config.ToolsConfig{Claude: true},
		Skills: []types.ResolvedSkill{{Name: "nonexistent-skill-xyz"}},
	}
	err := NewSkillInjector().InjectAll(rc)
	if err == nil {
		t.Fatal("expected error for missing local skill file, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-skill-xyz") {
		t.Errorf("error should mention skill name, got: %v", err)
	}
}

func TestSkillInjector_RemoveAll_RemovesBlock(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	claudeMD := filepath.Join(dir, "CLAUDE.md")
	content := "# Header\n\n<!-- usher:skill:supabase:start -->\nSkill content.\n<!-- usher:skill:supabase:end -->\n\n# Footer\n"
	if err := os.WriteFile(claudeMD, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rc := types.ResolvedConfig{Tools: config.ToolsConfig{Claude: true}}
	if err := NewSkillInjector().RemoveAll([]string{"supabase"}, rc); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	data, _ := os.ReadFile(claudeMD)
	got := string(data)
	if strings.Contains(got, "usher:skill:supabase") {
		t.Error("skill markers should have been removed")
	}
	if strings.Contains(got, "Skill content.") {
		t.Error("skill content should have been removed")
	}
	if !strings.Contains(got, "# Header") {
		t.Error("content before block was lost")
	}
	if !strings.Contains(got, "# Footer") {
		t.Error("content after block was lost")
	}
}
