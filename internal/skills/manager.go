package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstalledSkill represents a skill found on disk.
type InstalledSkill struct {
	Name        string
	Description string
	Path        string // master copy path
	Global      bool
}

// Manager handles install/remove/list of skills.
type Manager struct{}

func New() *Manager { return &Manager{} }

// Install clones owner/repo, finds all SKILL.md files, installs each as a skill.
func (m *Manager) Install(ownerRepo string, global bool) ([]InstalledSkill, error) {
	tmpDir, err := os.MkdirTemp("", "usher-skill-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := gitClone(ownerRepo, tmpDir); err != nil {
		return nil, err
	}

	skillFiles, err := findSkillFiles(tmpDir)
	if err != nil {
		return nil, err
	}
	if len(skillFiles) == 0 {
		return nil, fmt.Errorf("no SKILL.md files found in %s", ownerRepo)
	}

	base := masterBase(global)
	var installed []InstalledSkill
	for _, sf := range skillFiles {
		name := skillName(sf)
		dst := filepath.Join(base, name)

		if err := os.MkdirAll(dst, 0755); err != nil {
			return nil, err
		}
		if err := copyFile(sf, filepath.Join(dst, "SKILL.md")); err != nil {
			return nil, err
		}

		if err := createSymlinks(name, dst, global); err != nil {
			return nil, err
		}

		desc := readDescription(filepath.Join(dst, "SKILL.md"))
		installed = append(installed, InstalledSkill{
			Name:        name,
			Description: desc,
			Path:        dst,
			Global:      global,
		})
	}
	return installed, nil
}

// Remove deletes the master copy and all agent symlinks for skillName.
func (m *Manager) Remove(name string, global bool) error {
	base := masterBase(global)
	dst := filepath.Join(base, name)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not found (scope: %s)", name, scopeLabel(global))
	}

	// Remove symlinks from agent dirs first.
	for _, dir := range agentDirs {
		link := filepath.Join(expandHome(dir), name)
		_ = os.Remove(link)
	}
	return os.RemoveAll(dst)
}

// List returns all installed skills for the given scope.
func (m *Manager) List(global bool) ([]InstalledSkill, error) {
	base := masterBase(global)
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []InstalledSkill
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		skillPath := filepath.Join(base, e.Name())
		desc := readDescription(filepath.Join(skillPath, "SKILL.md"))
		result = append(result, InstalledSkill{
			Name:        e.Name(),
			Description: desc,
			Path:        skillPath,
			Global:      global,
		})
	}
	return result, nil
}

// FindByPrefix returns installed skill names that match name exactly or by prefix.
func (m *Manager) FindByPrefix(name string, global bool) ([]string, error) {
	skills, err := m.List(global)
	if err != nil {
		return nil, err
	}
	for _, s := range skills {
		if s.Name == name {
			return []string{name}, nil
		}
	}
	var matches []string
	for _, s := range skills {
		if strings.HasPrefix(s.Name, name) {
			matches = append(matches, s.Name)
		}
	}
	return matches, nil
}

func masterBase(global bool) string {
	if global {
		return expandHome(masterDir)
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, projectMasterDir)
}

func scopeLabel(global bool) string {
	if global {
		return "global"
	}
	return "project"
}

// createSymlinks creates symlinks from each global agent dir to the master copy.
// For project scope, no symlinks are needed — agents discover .agents/skills/ directly.
func createSymlinks(name, masterPath string, global bool) error {
	if !global {
		return nil
	}
	for _, dir := range agentDirs {
		agentSkillsDir := expandHome(dir)
		if err := os.MkdirAll(agentSkillsDir, 0755); err != nil {
			continue
		}
		link := filepath.Join(agentSkillsDir, name)
		_ = os.Remove(link)

		rel, err := filepath.Rel(agentSkillsDir, masterPath)
		if err != nil {
			rel = masterPath
		}
		_ = os.Symlink(rel, link)
	}
	return nil
}
