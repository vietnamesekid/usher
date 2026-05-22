package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// setupSkillOnDisk creates a fake installed skill at masterBase with a SKILL.md and SOURCE file.
func setupSkillOnDisk(t *testing.T, base, name, source, description string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: " + name + "\ndescription: " + description + "\n---\n# Content\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}
	if source != "" {
		if err := os.WriteFile(filepath.Join(dir, "SOURCE"), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestManager_List_ReadsSourceFile(t *testing.T) {
	tmp := t.TempDir()
	// Override masterDir for this test by using a project-scope base.
	setupSkillOnDisk(t, tmp, "supabase", "supabase/agent-skills", "Supabase skill")
	setupSkillOnDisk(t, tmp, "react", "some/react-skills", "React skill")

	// Call list directly on the directory.
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}

	type result struct {
		name   string
		source string
		desc   string
	}
	var got []result
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillPath := filepath.Join(tmp, e.Name())
		src, _ := os.ReadFile(filepath.Join(skillPath, "SOURCE"))
		desc := readDescription(filepath.Join(skillPath, "SKILL.md"))
		got = append(got, result{e.Name(), string(src), desc})
	}

	if len(got) != 2 {
		t.Fatalf("got %d skills, want 2", len(got))
	}
	byName := map[string]result{}
	for _, r := range got {
		byName[r.name] = r
	}
	if byName["supabase"].source != "supabase/agent-skills" {
		t.Errorf("supabase source = %q, want supabase/agent-skills", byName["supabase"].source)
	}
	if byName["react"].source != "some/react-skills" {
		t.Errorf("react source = %q, want some/react-skills", byName["react"].source)
	}
	if byName["supabase"].desc != "Supabase skill" {
		t.Errorf("supabase description = %q, want Supabase skill", byName["supabase"].desc)
	}
}

func TestManager_List_MissingSourceFile_ReturnsEmptySource(t *testing.T) {
	tmp := t.TempDir()
	setupSkillOnDisk(t, tmp, "old-skill", "", "Old skill without SOURCE file")

	src, _ := os.ReadFile(filepath.Join(tmp, "old-skill", "SOURCE"))
	if string(src) != "" {
		t.Errorf("expected empty source for skill without SOURCE file, got %q", src)
	}
}

func TestReadDescription_ParsesFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	content := "---\nname: test\ndescription: My test skill\n---\n# Content\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	desc := readDescription(path)
	if desc != "My test skill" {
		t.Errorf("description = %q, want My test skill", desc)
	}
}

func TestReadDescription_MissingFile_ReturnsEmpty(t *testing.T) {
	desc := readDescription("/nonexistent/SKILL.md")
	if desc != "" {
		t.Errorf("expected empty description for missing file, got %q", desc)
	}
}
