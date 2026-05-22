# Usher

**One command to rule all your AI coding tools.**

Usher installs and configures MCP servers and skills across Claude Code, Gemini CLI, Codex, and Cursor from a single place. Add a tool once — Usher writes the right config for every agent.

```bash
usher mcp add supabase      # stores token in Keychain, writes to all tools at once
usher skill add supabase    # installs skill globally for all agents
usher sync                  # re-sync after any config change
usher doctor                # verify everything is healthy
```

---

## Install

### Homebrew (macOS / Linux)

```bash
brew install vietnamesekid/tap/usher
```

### Download binary

Download the latest binary for your platform from [Releases](https://github.com/vietnamesekid/usher/releases):

| Platform | Architecture | Download |
| --- | --- | --- |
| macOS | Apple Silicon (M1/M2/M3) | `usher-darwin-arm64` |
| macOS | Intel | `usher-darwin-amd64` |
| Linux | x86_64 | `usher-linux-amd64` |
| Linux | ARM64 | `usher-linux-arm64` |
| Windows | x86_64 | `usher-windows-amd64.exe` |

```bash
# macOS example
curl -L https://github.com/vietnamesekid/usher/releases/latest/download/usher-darwin-arm64 \
  -o /usr/local/bin/usher && chmod +x /usr/local/bin/usher
```

### Go install

```bash
go install github.com/vietnamesekid/usher@latest
```

### Build from source

```bash
git clone https://github.com/vietnamesekid/usher
cd usher
make install
```

---

## Quickstart

```bash
# 1. Initialize — creates ~/.usher/config.json
usher setup

# 2. Add an MCP server (prompts for token, stores in OS Keychain)
usher mcp add supabase

# 3. Install a skill for all your AI agents
usher skill add supabase

# 4. Health check
usher doctor
```

---

## Commands

### MCP servers

```text
usher mcp add <name>            Add an MCP server and sync to all tools
usher mcp remove <name>         Remove an MCP server
usher mcp list                  List configured MCP servers
usher mcp list --available      Browse all MCPs in the registry
```

### Skills

```text
usher skill add <owner/skill>   Install a skill from skills.sh (e.g. supabase/agent-skills)
usher skill add supabase        Short name — Usher resolves to the right repo automatically
usher skill remove              Interactive: select from installed skills to remove
usher skill remove <name>       Remove a specific skill by name
usher skill update              Update all installed skills to latest
usher skill list                List installed skills
```

### Other

```text
usher setup                     Interactive setup wizard
usher sync                      Re-sync all tool configs
usher auth setup                Configure API keys for AI providers
usher auth revoke <provider>    Remove provider credentials
usher doctor                    Health check for all configured tools
```

---

## How MCP add works

```text
usher mcp add supabase
  │
  ├─ looks up "supabase" in the bundled registry (no network)
  ├─ prompts for SUPABASE_ACCESS_TOKEN
  ├─ stores token in OS Keychain (never written to disk)
  ├─ writes ~/.usher/config.json  ← keychain reference only, no token
  └─ syncs to all enabled tools:
       ~/.claude/settings.json    Claude Code
       ~/.gemini/settings.json    Gemini CLI
       ~/.codex/config.toml       Codex
       ~/.cursor/mcp.json         Cursor
```

Tokens are **never stored in config files** — only a keychain key reference. Syncing resolves tokens from the keychain at runtime.

## How skill add works

```text
usher skill add supabase
  │
  ├─ runs `npx skills find supabase` to resolve → "supabase/agent-skills"
  ├─ clones https://github.com/supabase/agent-skills (shallow)
  ├─ finds all SKILL.md files in the repo
  ├─ copies each to ~/.agents/skills/<skill-name>/  (global)
  │    or  .agents/skills/<skill-name>/             (project)
  └─ creates symlinks for each agent:
       ~/.claude/skills/<skill-name>  →  ../../.agents/skills/<skill-name>
       ~/.cursor/skills/<skill-name>  →  ../../.agents/skills/<skill-name>
       (and so on for all supported agents)
```

---

## Secrets storage

| Platform | Backend |
| --- | --- |
| macOS | Keychain Access (service: `usher`) |
| Linux | libsecret / GNOME Keyring |
| Windows | Windows Credential Manager |
| Headless / CI | `~/.usher/.secrets` (chmod 600) |

Force file-based fallback at any time:

```bash
USHER_KEYCHAIN_FALLBACK=1 usher mcp add supabase
```

---

## Project-level config

Add `.usher/project.json` to your repo to layer project-specific MCPs on top of your global config:

```json
{
  "mcpServers": {
    "supabase": {
      "instances": [
        {
          "name": "supabase-staging",
          "auth": { "type": "keychain", "key": "usher.mcp.supabase-staging.token" },
          "enabled": true
        }
      ]
    }
  }
}
```

Project config adds MCPs but never overrides global ones. Commit this file — tokens are never in it.

---

## MCP Registry

Usher ships with a bundled MCP registry (`internal/registry/mcp/*.json`). One file per MCP — no code changes needed to add one.

Currently bundled:

| Name | Description |
| --- | --- |
| `supabase` | Supabase database, auth, edge functions, storage |
| `github` | GitHub repos, issues, PRs, code search |
| `filesystem` | Read/write local files (no auth required) |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

MIT
