package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Writer struct {
	path string
}

func NewWriter(path string) *Writer {
	return &Writer{path: path}
}

func (w *Writer) AddMCPEntry(name string, entry MCPEntry) error {
	cfg, err := w.read()
	if err != nil {
		return err
	}
	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]MCPEntry)
	}
	cfg.MCPServers[name] = entry
	return w.write(cfg)
}

func (w *Writer) RemoveMCPEntry(name string) error {
	cfg, err := w.read()
	if err != nil {
		return err
	}
	delete(cfg.MCPServers, name)
	return w.write(cfg)
}

func (w *Writer) AddSkillEntry(name string, entry SkillEntry) error {
	cfg, err := w.read()
	if err != nil {
		return err
	}
	if cfg.Skills == nil {
		cfg.Skills = make(map[string]SkillEntry)
	}
	cfg.Skills[name] = entry
	return w.write(cfg)
}

func (w *Writer) RemoveSkillEntry(name string) error {
	cfg, err := w.read()
	if err != nil {
		return err
	}
	delete(cfg.Skills, name)
	return w.write(cfg)
}

func (w *Writer) AddAuthEntry(entry AuthEntry) error {
	cfg, err := w.read()
	if err != nil {
		return err
	}
	for i, a := range cfg.Auth {
		if a.Provider == entry.Provider {
			cfg.Auth[i] = entry
			return w.write(cfg)
		}
	}
	cfg.Auth = append(cfg.Auth, entry)
	return w.write(cfg)
}

func (w *Writer) SetTools(t ToolsConfig) error {
	cfg, err := w.read()
	if err != nil {
		return err
	}
	cfg.Tools = t
	return w.write(cfg)
}

func (w *Writer) Init(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0755); err != nil {
		return err
	}
	return w.write(cfg)
}

func (w *Writer) read() (Config, error) {
	data, err := os.ReadFile(w.path)
	if os.IsNotExist(err) {
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

func (w *Writer) write(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(w.path, data)
}

// atomicWrite writes data via a temp file + rename for crash safety.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
