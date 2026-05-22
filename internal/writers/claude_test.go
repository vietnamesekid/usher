package writers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/types"
)

func TestClaudeWriter_Write_InjectsMCPServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	rc := types.ResolvedConfig{
		Tools: config.ToolsConfig{Claude: true},
		MCPInstances: []types.ResolvedMCPInstance{
			{
				InstanceName: "supabase",
				Command:      "npx",
				Args:         []string{"-y", "@supabase/mcp-server-supabase@latest"},
				EnvVar:       "SUPABASE_ACCESS_TOKEN",
				Token:        "test-token-123",
			},
		},
	}

	if err := writeToPath(path, rc); err != nil {
		t.Fatalf("writeToPath: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	mcpRaw, ok := result["mcpServers"]
	if !ok {
		t.Fatal("mcpServers key missing from output")
	}

	var mcpServers map[string]claudeMCPServer
	if err := json.Unmarshal(mcpRaw, &mcpServers); err != nil {
		t.Fatalf("unmarshal mcpServers: %v", err)
	}

	srv, ok := mcpServers["supabase"]
	if !ok {
		t.Fatal("supabase MCP missing from mcpServers")
	}
	if srv.Command != "npx" {
		t.Errorf("command = %q, want npx", srv.Command)
	}
	if srv.Env["SUPABASE_ACCESS_TOKEN"] != "test-token-123" {
		t.Errorf("token = %q, want test-token-123", srv.Env["SUPABASE_ACCESS_TOKEN"])
	}
}

func TestClaudeWriter_Write_PreservesUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Write existing settings with unknown fields.
	existing := `{"theme":"dark","model":"sonnet","mcpServers":{}}`
	if err := os.WriteFile(path, []byte(existing), 0600); err != nil {
		t.Fatal(err)
	}

	rc := types.ResolvedConfig{
		Tools: config.ToolsConfig{Claude: true},
		MCPInstances: []types.ResolvedMCPInstance{
			{InstanceName: "github", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github@latest"}},
		},
	}

	if err := writeToPath(path, rc); err != nil {
		t.Fatalf("writeToPath: %v", err)
	}

	data, _ := os.ReadFile(path)
	var result map[string]json.RawMessage
	json.Unmarshal(data, &result)

	if _, ok := result["theme"]; !ok {
		t.Error("'theme' field was lost after write")
	}
	if _, ok := result["model"]; !ok {
		t.Error("'model' field was lost after write")
	}
	if _, ok := result["mcpServers"]; !ok {
		t.Error("'mcpServers' field missing after write")
	}
}

func TestClaudeWriter_Write_NoAuthMCP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	rc := types.ResolvedConfig{
		Tools: config.ToolsConfig{Claude: true},
		MCPInstances: []types.ResolvedMCPInstance{
			{
				InstanceName: "filesystem",
				Command:      "npx",
				Args:         []string{"-y", "@modelcontextprotocol/server-filesystem@latest"},
				EnvVar:       "", // no auth
				Token:        "",
			},
		},
	}

	if err := writeToPath(path, rc); err != nil {
		t.Fatalf("writeToPath: %v", err)
	}

	data, _ := os.ReadFile(path)
	var result map[string]json.RawMessage
	json.Unmarshal(data, &result)
	var mcpServers map[string]claudeMCPServer
	json.Unmarshal(result["mcpServers"], &mcpServers)

	srv := mcpServers["filesystem"]
	if len(srv.Env) != 0 {
		t.Errorf("no-auth MCP should have empty env, got %v", srv.Env)
	}
}

// writeToPath is a test helper that writes to an arbitrary path
// instead of the default ConfigPath.
func writeToPath(path string, rc types.ResolvedConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}
	mcpServers := make(map[string]claudeMCPServer)
	for _, inst := range rc.MCPInstances {
		srv := claudeMCPServer{Command: inst.Command, Args: inst.Args}
		if inst.EnvVar != "" && inst.Token != "" {
			srv.Env = map[string]string{inst.EnvVar: inst.Token}
		}
		mcpServers[inst.InstanceName] = srv
	}
	mcpJSON, err := json.Marshal(mcpServers)
	if err != nil {
		return err
	}
	raw["mcpServers"] = mcpJSON
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}
