package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vietnamesekid/usher/internal/types"
)

const (
	markerStartFmt = "<!-- usher:skill:%s:start -->"
	markerEndFmt   = "<!-- usher:skill:%s:end -->"
)

// SkillInjector injects skill content from local disk into instruction files.
type SkillInjector struct{}

func NewSkillInjector() *SkillInjector { return &SkillInjector{} }

// InjectAll processes all skills in rc, injecting into instruction files
// for each enabled tool.
func (si *SkillInjector) InjectAll(rc types.ResolvedConfig) error {
	files := instructionFiles(rc)
	for _, skill := range rc.Skills {
		content, err := readLocalSkill(skill.Name)
		if err != nil {
			return fmt.Errorf("reading skill %s: %w", skill.Name, err)
		}
		for _, f := range files {
			if err := injectIntoFile(f, skill.Name, content); err != nil {
				return fmt.Errorf("injecting skill %s into %s: %w", skill.Name, f, err)
			}
		}
	}
	return nil
}

// RemoveAll removes marker blocks for all named skills from instruction files.
func (si *SkillInjector) RemoveAll(skillNames []string, rc types.ResolvedConfig) error {
	files := instructionFiles(rc)
	for _, name := range skillNames {
		for _, f := range files {
			if err := removeFromFile(f, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// readLocalSkill reads SKILL.md from the global master install path.
func readLocalSkill(name string) (string, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".agents", "skills", name, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("skill %q not found at %s (run: usher skill add)", name, path)
	}
	return string(data), nil
}

// instructionFiles returns the instruction file paths for each enabled tool.
func instructionFiles(rc types.ResolvedConfig) []string {
	var files []string
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	if rc.Tools.Claude {
		files = append(files, cwd+"/CLAUDE.md")
	}
	if rc.Tools.Gemini {
		files = append(files, cwd+"/GEMINI.md")
	}
	if rc.Tools.Codex {
		files = append(files, cwd+"/AGENTS.md")
	}
	if rc.Tools.Cursor {
		files = append(files, cwd+"/.cursorrules")
	}
	if rc.Tools.Windsurf {
		files = append(files, home+"/.codeium/windsurf/memories/global_rules.md")
	}
	if rc.Tools.Cline {
		files = append(files, cwd+"/.clinerules")
	}
	return files
}

// injectIntoFile finds or creates the marker block for skillName in filePath.
// Content outside the markers is never modified.
func injectIntoFile(filePath, skillName, content string) error {
	start := fmt.Sprintf(markerStartFmt, skillName)
	end := fmt.Sprintf(markerEndFmt, skillName)
	block := start + "\n" + content + "\n" + end

	existing := ""
	if data, err := os.ReadFile(filePath); err == nil {
		existing = string(data)
	}

	startIdx := strings.Index(existing, start)
	endIdx := strings.Index(existing, end)

	if startIdx == -1 || endIdx == -1 {
		newContent := existing
		if newContent != "" && !strings.HasSuffix(newContent, "\n") {
			newContent += "\n"
		}
		newContent += "\n" + block + "\n"
		return os.WriteFile(filePath, []byte(newContent), 0644)
	}

	before := existing[:startIdx]
	after := existing[endIdx+len(end):]
	return os.WriteFile(filePath, []byte(before+block+after), 0644)
}

// removeFromFile removes the marker block for skillName from filePath.
func removeFromFile(filePath, skillName string) error {
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	start := fmt.Sprintf(markerStartFmt, skillName)
	end := fmt.Sprintf(markerEndFmt, skillName)
	content := string(data)

	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)
	if startIdx == -1 || endIdx == -1 {
		return nil
	}

	before := strings.TrimRight(content[:startIdx], "\n")
	after := strings.TrimLeft(content[endIdx+len(end):], "\n")
	result := before
	if after != "" {
		result += "\n" + after
	}
	return os.WriteFile(filePath, []byte(result), 0644)
}
