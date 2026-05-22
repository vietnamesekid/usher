package actions

import (
	"fmt"
	"os"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/keychain"
	"github.com/vietnamesekid/usher/internal/ui"
	"github.com/vietnamesekid/usher/internal/writers"
)

type DoctorActions struct {
	cfgLoader *config.Loader
	kc        keychain.Keychain
	out       *ui.Output
}

func NewDoctorActions(cfgLoader *config.Loader, kc keychain.Keychain, out *ui.Output) *DoctorActions {
	return &DoctorActions{cfgLoader: cfgLoader, kc: kc, out: out}
}

func (a *DoctorActions) Check() error {
	global, _, err := a.cfgLoader.Load()
	if err != nil {
		return err
	}

	headers := []string{"TOOL", "BINARY", "CONFIG", "MCPs", "NOTES"}
	var rows [][]string

	for _, w := range writers.All() {
		toolName := w.Name()
		if !isToolEnabledForDoctor(toolName, global.Tools) {
			continue
		}

		binaryOK := w.Detect()
		configOK := fileExists(w.ConfigPath())
		mcpCount := 0
		var notes []string

		if !binaryOK {
			notes = append(notes, binaryHint(toolName))
		}
		if !configOK {
			notes = append(notes, "run: usher sync")
		}

		for _, entry := range global.MCPServers {
			for _, inst := range entry.Instances {
				if !inst.Enabled {
					continue
				}
				ok, _ := a.kc.Exists(inst.Auth.Key)
				if ok {
					mcpCount++
				} else {
					notes = append(notes, fmt.Sprintf("missing token: %s", inst.Auth.Key))
				}
			}
		}

		noteStr := "ok"
		if len(notes) > 0 {
			noteStr = notes[0]
		}

		rows = append(rows, []string{
			toolName,
			checkmark(binaryOK),
			checkmark(configOK),
			fmt.Sprintf("%d", mcpCount),
			noteStr,
		})
	}

	a.out.Table(headers, rows)
	return nil
}

func checkmark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func binaryHint(toolName string) string {
	hints := map[string]string{
		"claude": "npm install -g @anthropic-ai/claude-code",
		"gemini": "npm install -g @google/gemini-cli",
		"codex":  "npm install -g @openai/codex",
		"cursor": "download at cursor.sh",
	}
	if hint, ok := hints[toolName]; ok {
		return fmt.Sprintf("binary not found → %s", hint)
	}
	return "binary not found"
}

func isToolEnabledForDoctor(name string, tools config.ToolsConfig) bool {
	switch name {
	case "claude":
		return tools.Claude
	case "gemini":
		return tools.Gemini
	case "codex":
		return tools.Codex
	case "cursor":
		return tools.Cursor
	}
	return false
}
