package skills

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitClone clones github.com/ownerRepo into destDir (shallow clone).
func gitClone(ownerRepo, destDir string) error {
	url := fmt.Sprintf("https://github.com/%s.git", ownerRepo)
	cmd := exec.Command("git", "clone", "--depth=1", url, destDir)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cloning %s: %w", ownerRepo, err)
	}
	return nil
}

// findSkillFiles walks dir and returns all SKILL.md paths found.
func findSkillFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.EqualFold(info.Name(), "SKILL.md") {
			// Skip the repo root SKILL.md if it's a meta/template.
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// skillName derives the skill name from the SKILL.md path.
// Uses the parent directory name (e.g. .../supabase/SKILL.md → "supabase").
// Falls back to the frontmatter `name:` field.
func skillName(skillPath string) string {
	dir := filepath.Dir(skillPath)
	base := filepath.Base(dir)
	// If parent dir is the cloned repo root (contains .git), use frontmatter name.
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		if name := readFrontmatterField(skillPath, "name"); name != "" {
			return name
		}
	}
	return base
}

// readDescription reads the `description:` field from SKILL.md frontmatter.
func readDescription(skillPath string) string {
	return readFrontmatterField(skillPath, "description")
}

// readFrontmatterField reads a YAML frontmatter field value from a SKILL.md.
func readFrontmatterField(path, field string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	prefix := field + ":"
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if inFrontmatter && strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// copyFile copies src to dst, creating dst's parent dirs.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
