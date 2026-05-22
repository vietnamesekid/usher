# Contributing to Usher

Thanks for your interest in contributing! Usher is designed so most contributions touch **one file only**.

## Ways to contribute

| Contribution | Files changed |
| --- | --- |
| Add a new MCP server | 1 JSON file |
| Add a custom skill | 1 JSON file |
| Add support for a new AI tool | 3 files |
| Bug fix / improvement | varies |

---

## Add a new MCP server

Create `internal/registry/mcp/{name}.json`:

```json
{
  "name": "yourserver",
  "description": "One sentence about what it does",
  "command": "npx",
  "args": ["-y", "@your-org/mcp-server-yourserver@latest"],
  "auth": {
    "envVar": "YOURSERVER_API_KEY",
    "url": "https://yourserver.com/settings/tokens"
  },
  "tags": ["category"]
}
```

For MCPs with no authentication (like `filesystem`), use `"envVar": ""` and `"url": ""`.

The `//go:embed mcp/*.json` directive in `internal/registry/embed.go` picks up the new file automatically — no Go code changes needed. Open a PR with just this one file.

**Validate your JSON before opening a PR:**

```bash
make validate-registry
```

---

## Add a custom skill

Create `internal/registry/skills/{name}.json` for skills you host yourself (URL or local):

```json
{
  "name": "yourskill",
  "description": "What this skill teaches the AI",
  "source": {
    "type": "url",
    "url": "https://raw.githubusercontent.com/your-org/skills/main/{name}/{version}.md"
  },
  "versions": ["1.0.0"],
  "latest": "1.0.0",
  "tags": ["category"]
}
```

> Skills published on [skills.sh](https://skills.sh) don't need a registry entry — users install them directly with `usher skill add owner/repo`.

---

## Add support for a new AI tool

**Step 1** — Create `internal/writers/{toolname}.go` implementing the `Writer` interface:

```go
type Writer interface {
    Name() string                         // e.g. "zed"
    Detect() bool                         // is the binary / config dir present?
    ConfigPath() string                   // absolute path to the tool's config file
    Backup(backupsDir string) error       // copy current config before overwriting
    Write(rc types.ResolvedConfig) error  // generate and write the new config
}
```

Use `internal/writers/claude.go` as a reference for JSON-based configs, or `internal/writers/codex.go` for TOML.

**Step 2** — Register the writer in `internal/writers/writer.go`:

```go
func All() []Writer {
    return []Writer{
        NewClaudeWriter(),
        NewGeminiWriter(),
        NewCodexWriter(),
        NewCursorWriter(),
        NewYourToolWriter(), // ← add here
    }
}
```

**Step 3** — Add a field to `ToolsConfig` in `internal/config/config.go`:

```go
type ToolsConfig struct {
    Claude bool `json:"claude"`
    Gemini bool `json:"gemini"`
    Codex  bool `json:"codex"`
    Cursor bool `json:"cursor"`
    YourTool bool `json:"yourtool"` // ← add here
}
```

The sync engine, actions, and all other writers need no changes.

---

## Development setup

```bash
git clone https://github.com/vietnamesekid/usher
cd usher
go mod download
make build     # builds ./usher binary
make test      # runs all tests with race detector
make lint      # requires golangci-lint
```

### Run a specific test

```bash
go test ./internal/registry/... -v -run TestGetMCP_Found
```

### Project layout

```
cmd/              Cobra commands — thin wrappers, no logic
internal/
  actions/        Business logic (mcp.go, skill.go, auth.go, doctor.go)
  config/         Config types, loader, writer
  keychain/       OS keychain + file fallback
  registry/       Embedded JSON registry + fuzzy lookup
  skills/         Skill clone/install/remove/list
  sync/           Orchestration: merge → resolve → write
  types/          Shared types (ResolvedConfig) — no import cycles
  ui/             Output formatting and interactive prompts
  writers/        Per-tool config writers
```

---

## Pull request guidelines

- Keep PRs focused — one feature or fix per PR
- Add tests for new behaviour in the relevant `*_test.go` file
- Run `make test` and `make validate-registry` before opening a PR
- Commit messages: imperative present tense (`Add supabase MCP`, not `Added`)

---

## Reporting issues

Open an issue at [github.com/vietnamesekid/usher/issues](https://github.com/vietnamesekid/usher/issues).

Include:
- OS and version
- `usher --version` output
- The command you ran
- Full error output
