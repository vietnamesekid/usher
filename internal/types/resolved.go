package types

import (
	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/registry"
)

// ResolvedMCPInstance is a single MCP server with its token resolved from keychain.
// The Token field MUST NOT be serialized to disk.
type ResolvedMCPInstance struct {
	InstanceName string
	Command      string
	Args         []string
	EnvVar       string
	Token        string // plaintext token — NEVER persist this
}

// ResolvedSkill is a configured skill ready for injection.
type ResolvedSkill struct {
	Name    string
	Version string
	Source  registry.SkillSource
}

// ResolvedConfig is the fully-expanded, secret-populated config.
// It MUST NOT be serialized to disk in any form.
type ResolvedConfig struct {
	MCPInstances []ResolvedMCPInstance
	Skills       []ResolvedSkill
	Tools        config.ToolsConfig
	BackupsDir   string
}
