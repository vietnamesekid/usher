package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Loader struct {
	globalPath  string
	projectPath string
}

func NewLoader(globalOverride string) *Loader {
	l := &Loader{}
	if globalOverride != "" {
		l.globalPath = globalOverride
	} else {
		l.globalPath = defaultGlobalPath()
	}
	l.projectPath = defaultProjectPath()
	return l
}

// Load reads global + optional project config and returns the merged result.
// The second return value indicates whether a project config was found.
func (l *Loader) Load() (Config, bool, error) {
	global, err := l.LoadGlobal()
	if err != nil {
		return Config{}, false, err
	}
	project, found, err := l.LoadProject()
	if err != nil {
		return Config{}, false, err
	}
	if !found {
		return global, false, nil
	}
	return mergeConfigs(global, project), true, nil
}

func (l *Loader) LoadGlobal() (Config, error) {
	data, err := os.ReadFile(l.globalPath)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]MCPEntry)
	}
	if cfg.Skills == nil {
		cfg.Skills = make(map[string]SkillEntry)
	}
	return cfg, nil
}

func (l *Loader) LoadProject() (ProjectConfig, bool, error) {
	data, err := os.ReadFile(l.projectPath)
	if errors.Is(err, os.ErrNotExist) {
		return ProjectConfig{}, false, nil
	}
	if err != nil {
		return ProjectConfig{}, false, err
	}
	var proj ProjectConfig
	if err := json.Unmarshal(data, &proj); err != nil {
		return ProjectConfig{}, false, err
	}
	return proj, true, nil
}

// LoadBoth returns the global config and project config separately,
// for callers that need to pass them individually to sync.
func (l *Loader) LoadBoth() (Config, ProjectConfig, error) {
	global, err := l.LoadGlobal()
	if err != nil {
		return Config{}, ProjectConfig{}, err
	}
	project, _, err := l.LoadProject()
	return global, project, err
}

func (l *Loader) GlobalPath() string  { return l.globalPath }
func (l *Loader) ProjectPath() string { return l.projectPath }

// mergeConfigs combines global and project configs.
// MCPServers: union (project adds to global, does not override).
// Skills: union, project entries with Disabled=true are removed from result.
// Tools and Auth: from global only.
func mergeConfigs(global Config, project ProjectConfig) Config {
	result := global

	result.MCPServers = make(map[string]MCPEntry)
	for k, v := range global.MCPServers {
		result.MCPServers[k] = v
	}
	for k, v := range project.MCPServers {
		if _, exists := result.MCPServers[k]; !exists {
			result.MCPServers[k] = v
		}
	}

	result.Skills = make(map[string]SkillEntry)
	for k, v := range global.Skills {
		result.Skills[k] = v
	}
	for k, v := range project.Skills {
		if v.Disabled {
			delete(result.Skills, k)
		} else if _, exists := result.Skills[k]; !exists {
			result.Skills[k] = v
		}
	}

	return result
}

func defaultGlobalPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".usher/config.json"
	}
	return filepath.Join(home, ".usher", "config.json")
}

func defaultProjectPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ".usher/project.json"
	}
	return filepath.Join(cwd, ".usher", "project.json")
}
