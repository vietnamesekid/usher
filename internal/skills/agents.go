package skills

import "os"

// agentDirs maps agent name → skill directory path (global scope).
// These match the directories used by skills.sh.
var agentDirs = map[string]string{
	"claude":  "~/.claude/skills",
	"cursor":  "~/.cursor/skills",
	"codex":   "~/.codex/skills",
	"gemini":  "~/.gemini/skills",
	"windsurf": "~/.windsurf/skills",
	"cline":   "~/.vscode/extensions/saoudrizwan.claude-dev/skills",
}

// masterDir is where skill content is stored; agent dirs symlink here.
const masterDir = "~/.agents/skills"
const projectMasterDir = ".agents/skills"

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}
