# Usher

Configure MCP servers and skills once. Usher syncs them to every AI coding tool you use.

Usher manages [MCP servers](https://modelcontextprotocol.io) and agent skills across Claude Code, Gemini CLI, Codex, Cursor, Windsurf, and Cline from a single config file. Add a tool once and Usher writes the right format for every agent.

```bash
usher mcp add supabase      # store token in Keychain, write MCP config to all agents
usher skill add supabase    # install skill globally across all agents
usher sync                  # re-sync after any config change
usher doctor                # verify everything is healthy
```

---

## Table of Contents

- [Install](#install)
- [Quickstart](#quickstart)
- [Commands](#commands)
- [How it works](#how-it-works)
- [Secrets storage](#secrets-storage)
- [Project-level config](#project-level-config)
- [MCP registry](#mcp-registry)
- [Supported agents](#supported-agents)
- [Why Usher exists](#why-usher-exists)
- [Contributing](#contributing)

---

## Install

### Homebrew

```bash
brew install vietnamesekid/tap/usher
```

### Download binary

| Platform | Architecture | File |
| --- | --- | --- |
| macOS | Apple Silicon | `usher-darwin-arm64` |
| macOS | Intel | `usher-darwin-amd64` |
| Linux | x86_64 | `usher-linux-amd64` |
| Linux | ARM64 | `usher-linux-arm64` |
| Windows | x86_64 | `usher-windows-amd64.exe` |
| Windows | ARM64 | `usher-windows-arm64.exe` |

Download from the [Releases page](https://github.com/vietnamesekid/usher/releases), then make it executable:

```bash
# macOS / Linux
curl -L https://github.com/vietnamesekid/usher/releases/latest/download/usher-darwin-arm64 \
  -o /usr/local/bin/usher
chmod +x /usr/local/bin/usher
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
# 1. Run the setup wizard, creates ~/.usher/config.json
usher setup

# 2. Add an MCP server
usher mcp add supabase

# 3. Install a skill for all your AI agents
usher skill add supabase

# 4. Verify everything is healthy
usher doctor
```

---

## Commands

### MCP servers

| Command | Description |
| --- | --- |
| `usher mcp add <name>` | Add an MCP server and sync to all tools |
| `usher mcp remove <name>` | Remove an MCP server |
| `usher mcp list` | List configured MCP servers |
| `usher mcp list --available` | Browse all MCPs in the registry |

### Skills

| Command | Description |
| --- | --- |
| `usher skill add <owner/repo>` | Install a skill from GitHub (e.g. `supabase/agent-skills`) |
| `usher skill add <name>` | Short name, Usher resolves to the right repo automatically |
| `usher skill remove` | Interactive prompt to select and remove an installed skill |
| `usher skill remove <name>` | Remove a specific skill by name |
| `usher skill update` | Update all installed skills to latest |
| `usher skill list` | List installed skills |

### Other

| Command | Description |
| --- | --- |
| `usher setup` | Interactive setup wizard |
| `usher sync` | Re-sync all tool configs |
| `usher auth setup` | Configure API keys for AI providers |
| `usher auth revoke <provider>` | Remove provider credentials |
| `usher doctor` | Health check for all configured tools |

---

## How it works

### MCP add

```text
usher mcp add supabase
  |
  +-- looks up "supabase" in the bundled registry (no network required)
  +-- prompts for SUPABASE_ACCESS_TOKEN
  +-- stores the token in the OS Keychain (never written to disk)
  +-- writes ~/.usher/config.json with a keychain reference, not the token
  +-- syncs to all enabled tools:
        ~/.claude/settings.json                                                              (Claude Code)
        ~/.gemini/settings.json                                                              (Gemini CLI)
        ~/.codex/config.toml                                                                 (Codex)
        ~/.cursor/mcp.json                                                                   (Cursor)
        ~/.codeium/windsurf/mcp_config.json                                                  (Windsurf)
        ~/Library/Application Support/Code/User/globalStorage/.../cline_mcp_settings.json   (Cline)
```

Tokens are **never stored in config files**. Each sync resolves them from the keychain at runtime.

### Skill add

```text
usher skill add supabase
  |
  +-- resolves "supabase" to "supabase/agent-skills" via npx skills find
  +-- shallow-clones https://github.com/supabase/agent-skills into a temp dir
  +-- finds all SKILL.md files in the repo
  +-- copies each skill to:
        ~/.agents/skills/<skill-name>/        (global scope)
        .agents/skills/<skill-name>/          (project scope)
  +-- creates a symlink from each agent's skills directory to the master copy:
        ~/.claude/skills/<skill-name>
        ~/.gemini/skills/<skill-name>
        ~/.codex/skills/<skill-name>
        ~/.cursor/skills/<skill-name>
        ~/.windsurf/skills/<skill-name>
        ~/.vscode/extensions/saoudrizwan.claude-dev/skills/<skill-name>
  +-- injects skill content into instruction files:
        CLAUDE.md, GEMINI.md, AGENTS.md, .cursorrules
```

Skill content is injected between marker comments and updated in-place on every `usher sync`:

```html
<!-- usher:skill:supabase:start -->
...skill content...
<!-- usher:skill:supabase:end -->
```

---

## Secrets storage

| Platform | Backend |
| --- | --- |
| macOS | Keychain Access (service: `usher`) |
| Linux | libsecret / GNOME Keyring |
| Windows | Windows Credential Manager |
| Headless / CI | `~/.usher/.secrets` (mode 600) |

To force the file-based fallback:

```bash
USHER_KEYCHAIN_FALLBACK=1 usher mcp add supabase
```

---

## Project-level config

Add `.usher/project.json` to your repo to layer project-specific MCPs on top of the global config:

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

Project config adds MCPs but never overrides global ones. This file is safe to commit — it contains only keychain key references, never tokens.

---

## MCP registry

Usher ships with a bundled registry at `internal/registry/mcp/*.json`. Each MCP is one JSON file — no Go code changes are needed to add a new entry.

| Name | Description |
| --- | --- |
| `supabase` | Supabase database, auth, edge functions, storage |
| `github` | GitHub repos, issues, PRs, code search |
| `filesystem` | Read/write local files (no auth required) |
| `postgres` | Query, inspect schema, and manage Postgres databases |
| `slack` | Read channels, send messages, manage workspaces |

To add a new MCP server, see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Supported agents

| Agent | MCP config | Skill directory | Instruction file |
| --- | --- | --- | --- |
| Claude Code | `~/.claude/settings.json` | `~/.claude/skills/` | `CLAUDE.md` |
| Gemini CLI | `~/.gemini/settings.json` | `~/.gemini/skills/` | `GEMINI.md` |
| Codex | `~/.codex/config.toml` | `~/.codex/skills/` | `AGENTS.md` |
| Cursor | `~/.cursor/mcp.json` | `~/.cursor/skills/` | `.cursorrules` |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` | `~/.windsurf/skills/` | `~/.codeium/windsurf/memories/global_rules.md` |
| Cline | `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json` | `~/.vscode/extensions/saoudrizwan.claude-dev/skills/` | `.clinerules` |

---

## Why Usher exists

The AI coding tool ecosystem has fragmented fast. Most teams now run two or three agents side by side: Claude Code, Cursor, Gemini CLI, Codex, and every one of them has its own config format, its own location for MCP servers, and its own instruction file convention.

Without a coordinator, the pain compounds:

- **Config drift.** You add the Supabase MCP to Claude Code. Three weeks later a teammate enables Cursor. The Supabase MCP is not there. Someone adds it manually in a slightly different format. Now you have two sources of truth.
- **Secret sprawl.** Each tool has a different answer for where to store the access token. Some write it to a config file that ends up in git. Some write it to a dotfile in the home directory. One team member pastes the token into `.cursor/mcp.json` and commits it.
- **Skill rot.** Skill content is copy-pasted into `CLAUDE.md`, `AGENTS.md`, and `.cursorrules` separately. When the upstream skill updates, none of those copies get the change.
- **Onboarding friction.** A new developer needs to manually configure four tools before they can get started. The steps are not documented anywhere because they live in each person's head.

Usher treats the problem the way Homebrew treats package management: it does not create the tools, it knows how to configure them correctly. One config file, one command to sync, one place to add a new MCP server.

**Core design decisions:**

- **Secrets never touch disk.** Tokens are stored in the OS Keychain (macOS Keychain, libsecret, Windows Credential Manager). Config files only contain a keychain key reference. A config file is always safe to commit.
- **Writers are isolated.** Each agent has exactly one file responsible for its config format. Adding support for a new agent means adding one file, nothing else changes.
- **Sync is the only write path.** Nothing writes directly to a tool's config file. Every change goes through the sync pipeline: merge global and project config, resolve secrets from keychain, then write each tool's config from scratch. This guarantees all tools stay in sync.
- **Skills are managed, not copied.** Skill content lives in one place (`~/.agents/skills/`). Agent directories get a symlink. Instruction files get an injected block between comment markers, updated in-place on every sync.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE)
