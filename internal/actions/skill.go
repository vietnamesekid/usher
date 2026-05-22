package actions

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/skills"
	"github.com/vietnamesekid/usher/internal/ui"
)

type SkillActions struct {
	mgr       *skills.Manager
	cfgWriter *config.Writer
	syncFn    func(global config.Config, project config.ProjectConfig) error
	out       *ui.Output
	prompt    *ui.Prompt
}

func NewSkillActions(cfgWriter *config.Writer, syncFn func(config.Config, config.ProjectConfig) error, out *ui.Output, prompt *ui.Prompt) *SkillActions {
	return &SkillActions{
		mgr:       skills.New(),
		cfgWriter: cfgWriter,
		syncFn:    syncFn,
		out:       out,
		prompt:    prompt,
	}
}

// Add resolves slug, prompts scope if needed, clones and installs, then writes to config.
func (a *SkillActions) Add(slug string, globalFlag bool, globalFlagSet bool, agents []string, all bool) error {
	resolved, err := a.resolveSlug(slug)
	if err != nil {
		return err
	}

	global := globalFlag
	if !globalFlagSet {
		scope := a.prompt.AskSelect(
			"Install scope:",
			[]string{"global (~/.agents/skills, ...)", "project (current directory only)"},
		)
		global = strings.HasPrefix(scope, "global")
	}

	a.out.Step(fmt.Sprintf("Cloning %s...", resolved))
	installed, err := a.mgr.Install(resolved, global)
	if err != nil {
		return err
	}

	for _, s := range installed {
		entry := config.SkillEntry{
			Source: resolved,
		}
		if err := a.cfgWriter.AddSkillEntry(s.Name, entry); err != nil {
			return fmt.Errorf("saving skill %q to config: %w", s.Name, err)
		}
		a.out.Success(fmt.Sprintf("Installed skill %q → %s", s.Name, s.Path))
		if s.Description != "" {
			a.out.Info(fmt.Sprintf("  %s", s.Description))
		}
	}
	return nil
}

// Remove resolves name to installed skill(s), removes from disk and config.
func (a *SkillActions) Remove(name string, globalFlag bool, globalFlagSet bool, agents []string) error {
	global := globalFlag
	if !globalFlagSet {
		scope := a.prompt.AskSelect(
			"Remove from which scope?",
			[]string{"global (~/.agents/skills, ...)", "project (current directory only)"},
		)
		global = strings.HasPrefix(scope, "global")
	}

	var toRemove []string

	if name == "" {
		installed, err := a.mgr.List(global)
		if err != nil {
			return err
		}
		if len(installed) == 0 {
			a.out.Info("No skills installed in this scope.")
			return nil
		}
		names := make([]string, len(installed))
		for i, s := range installed {
			names[i] = s.Name
		}
		toRemove = a.prompt.AskMultiSelect("Select skills to remove:", names)
		if len(toRemove) == 0 {
			return nil
		}
	} else {
		targets, err := a.mgr.FindByPrefix(name, global)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return fmt.Errorf("no installed skill found matching %q\nRun `usher skill list` to see installed skills", name)
		}
		if len(targets) > 1 {
			toRemove = a.prompt.AskMultiSelect(
				fmt.Sprintf("Multiple skills match %q, select to remove:", name),
				targets,
			)
		} else {
			toRemove = targets
		}
	}

	for _, target := range toRemove {
		if err := a.mgr.Remove(target, global); err != nil {
			return err
		}
		if err := a.cfgWriter.RemoveSkillEntry(target); err != nil {
			return fmt.Errorf("removing skill %q from config: %w", target, err)
		}
		a.out.Success(fmt.Sprintf("Removed skill %q", target))
	}
	return nil
}

// Update re-installs all skills from their recorded source repos.
func (a *SkillActions) Update(globalFlag bool, globalFlagSet bool, agents []string, all bool) error {
	global := globalFlag
	if !globalFlagSet {
		scope := a.prompt.AskSelect(
			"Update which scope?",
			[]string{"global (~/.agents/skills, ...)", "project (current directory only)"},
		)
		global = strings.HasPrefix(scope, "global")
	}

	installed, err := a.mgr.List(global)
	if err != nil {
		return err
	}
	if len(installed) == 0 {
		a.out.Info("No skills installed in this scope.")
		return nil
	}

	// Re-install each skill from its recorded source repo.
	for _, s := range installed {
		source := s.Source
		if source == "" {
			a.out.Warning(fmt.Sprintf("Skipping %q: source repo not recorded (manually installed?)", s.Name))
			continue
		}
		a.out.Step(fmt.Sprintf("Updating %q from %s...", s.Name, source))
		if _, err := a.mgr.Install(source, global); err != nil {
			a.out.Warning(fmt.Sprintf("Failed to update %q: %v", s.Name, err))
			continue
		}
		a.out.Success(fmt.Sprintf("Updated %q", s.Name))
	}
	return nil
}

// List prints installed skills using usher's output format.
func (a *SkillActions) List() error {
	global, err := a.mgr.List(true)
	if err != nil {
		return err
	}
	project, err := a.mgr.List(false)
	if err != nil {
		return err
	}

	if len(global) == 0 && len(project) == 0 {
		a.out.Info("No skills installed. Run `usher skill add <owner/skill>` to install one.")
		return nil
	}

	if len(global) > 0 {
		a.out.Info("Global")
		rows := make([][]string, len(global))
		for i, s := range global {
			rows[i] = []string{s.Name, s.Source, truncate(s.Description, 60)}
		}
		a.out.Table([]string{"NAME", "SOURCE", "DESCRIPTION"}, rows)
	}
	if len(project) > 0 {
		if len(global) > 0 {
			a.out.Info("")
		}
		a.out.Info("Project")
		rows := make([][]string, len(project))
		for i, s := range project {
			rows[i] = []string{s.Name, s.Source, truncate(s.Description, 60)}
		}
		a.out.Table([]string{"NAME", "SOURCE", "DESCRIPTION"}, rows)
	}
	return nil
}

func truncate(s string, max int) string {
	s = strings.Trim(s, `"`)
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// resolveSlug resolves a short name like "supabase" to "supabase/agent-skills"
// by running `npx skills find <name>` and picking the best match.
// Slugs that already contain "/" are returned as-is.
func (a *SkillActions) resolveSlug(slug string) (string, error) {
	if strings.Contains(slug, "/") {
		return slug, nil
	}

	a.out.Step(fmt.Sprintf("Searching skills.sh for %q...", slug))
	out, err := exec.Command("npx", "skills", "find", slug).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("npx skills find %s: %w", slug, err)
	}

	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean := ansi.ReplaceAllString(string(out), "")

	var candidates []string
	seen := map[string]bool{}
	for _, line := range strings.Split(clean, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "@"); idx > 0 {
			ownerRepo := line[:idx]
			if strings.Contains(ownerRepo, "/") && !seen[ownerRepo] {
				seen[ownerRepo] = true
				candidates = append(candidates, ownerRepo)
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no skills found for %q on skills.sh", slug)
	}

	for _, c := range candidates {
		owner := strings.SplitN(c, "/", 2)[0]
		if strings.EqualFold(owner, slug) {
			return c, nil
		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	chosen := a.prompt.AskSelect(
		fmt.Sprintf("Multiple repos found for %q, pick one:", slug),
		candidates,
	)
	return chosen, nil
}
