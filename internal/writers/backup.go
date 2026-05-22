package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// backupFile copies src to backupsDir/{toolName}/{timestamp}{ext}.
// Returns nil if src does not exist (nothing to back up).
func backupFile(src, backupsDir, toolName string) error {
	data, err := os.ReadFile(src)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	dir := filepath.Join(backupsDir, toolName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	ext := filepath.Ext(src)
	ts := time.Now().Format("20060102-150405")
	dst := filepath.Join(dir, fmt.Sprintf("%s%s", ts, ext))
	return os.WriteFile(dst, data, 0600)
}
