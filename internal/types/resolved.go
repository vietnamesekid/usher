package types

import "github.com/vietnamesekid/usher/internal/config"

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
// Source is "owner/repo" — the local SKILL.md is read from ~/.agents/skills/<Name>/SKILL.md.
type ResolvedSkill struct {
	Name   string
	Source string // "owner/repo"
}

// ResolvedConfig is the fully-expanded, secret-populated config.
// It MUST NOT be serialized to disk in any form.
type ResolvedConfig struct {
	MCPInstances []ResolvedMCPInstance
	Skills       []ResolvedSkill
	Tools        config.ToolsConfig
	BackupsDir   string
}
