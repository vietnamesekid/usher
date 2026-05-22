package writers

import "github.com/vietnamesekid/usher/internal/types"

// Writer generates and writes the config file for a single AI tool.
type Writer interface {
	// Name returns the tool identifier (e.g. "claude", "gemini").
	Name() string
	// Detect returns true if the tool binary exists in PATH.
	Detect() bool
	// ConfigPath returns the absolute path to the tool's config file.
	ConfigPath() string
	// Backup copies the current config file to backupsDir/{tool}/{timestamp}.
	Backup(backupsDir string) error
	// Write generates and writes the tool config from the resolved config.
	Write(rc types.ResolvedConfig) error
}

// All returns every registered writer.
// To add support for a new tool: implement Writer in a new file,
// then add NewYourToolWriter() here.
func All() []Writer {
	return []Writer{
		NewClaudeWriter(),
		NewGeminiWriter(),
		NewCodexWriter(),
		NewCursorWriter(),
	}
}
