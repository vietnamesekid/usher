package sync

import (
	"fmt"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/registry"
)

// MergedConfig is the fully expanded intermediate representation
// before secrets are resolved. Not safe to store on disk.
type MergedConfig struct {
	Instances  []MergedMCPInstance
	Skills     []MergedSkill
	Tools      config.ToolsConfig
	BackupsDir string
}

type MergedMCPInstance struct {
	InstanceName string
	Command      string
	Args         []string
	EnvVar       string
	AuthKey      string // keychain key reference (not the plaintext token)
}

type MergedSkill struct {
	Name    string
	Version string
	Source  registry.SkillSource
}

// Merge combines global + project configs and expands each MCP entry
// using the registry to resolve command/args/envVar.
func Merge(global config.Config, project config.ProjectConfig, reg registry.Registry) (MergedConfig, error) {
	merged := config.Config{}
	// Re-use config loader merge logic via a direct call.
	merged.Tools = global.Tools
	merged.Sync = global.Sync

	merged.MCPServers = make(map[string]config.MCPEntry)
	for k, v := range global.MCPServers {
		merged.MCPServers[k] = v
	}
	for k, v := range project.MCPServers {
		if _, exists := merged.MCPServers[k]; !exists {
			merged.MCPServers[k] = v
		}
	}

	merged.Skills = make(map[string]config.SkillEntry)
	for k, v := range global.Skills {
		merged.Skills[k] = v
	}
	for k, v := range project.Skills {
		if v.Disabled {
			delete(merged.Skills, k)
		} else if _, exists := merged.Skills[k]; !exists {
			merged.Skills[k] = v
		}
	}

	mc := MergedConfig{
		Tools:      merged.Tools,
		BackupsDir: backupsDir(merged),
	}

	for serverName, entry := range merged.MCPServers {
		regEntry, err := reg.GetMCP(serverName)
		if err != nil {
			return MergedConfig{}, fmt.Errorf("registry lookup for %q: %w", serverName, err)
		}
		for _, inst := range entry.Instances {
			if !inst.Enabled {
				continue
			}
			mc.Instances = append(mc.Instances, MergedMCPInstance{
				InstanceName: inst.Name,
				Command:      regEntry.Command,
				Args:         regEntry.Args,
				EnvVar:       regEntry.Auth.EnvVar,
				AuthKey:      inst.Auth.Key,
			})
		}
	}

	for skillName, entry := range merged.Skills {
		if entry.Disabled {
			continue
		}
		regSkill, err := reg.GetSkill(skillName)
		if err != nil {
			return MergedConfig{}, fmt.Errorf("registry lookup for skill %q: %w", skillName, err)
		}
		version := entry.Version
		if version == "" {
			version = regSkill.Latest
		}
		mc.Skills = append(mc.Skills, MergedSkill{
			Name:    skillName,
			Version: version,
			Source:  regSkill.Source,
		})
	}

	return mc, nil
}

func backupsDir(cfg config.Config) string {
	if cfg.Sync.BackupsDir != "" {
		return cfg.Sync.BackupsDir
	}
	return "~/.usher/backups"
}
