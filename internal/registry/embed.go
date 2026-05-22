package registry

import "embed"

//go:embed mcp/*.json
var mcpFS embed.FS

//go:embed all:skills
var skillsFS embed.FS
