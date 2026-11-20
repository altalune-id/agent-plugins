# agent-plugins

[![CI](https://github.com/altalune-id/agent-plugins/actions/workflows/ci.yml/badge.svg)](https://github.com/altalune-id/agent-plugins/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Claude Code](https://img.shields.io/badge/Claude_Code-plugin-D97757)
![Codex](https://img.shields.io/badge/Codex-plugin-412991)
![OpenCode](https://img.shields.io/badge/OpenCode-skills-222222)
![Gemini CLI](https://img.shields.io/badge/Gemini_CLI-skills-4285F4)
![Cursor](https://img.shields.io/badge/Cursor-plugin-000000)

Shared skills, role agents, hooks and MCP servers for Go, Vue and htmx work. One source of skills; every tool reads the same folders. Upstream skills are pinned, and a weekly job opens a PR when a newer version exists.

## Plugins

| Plugin | What you get |
| --- | --- |
| **alt-workflow** | lead, design, spec and plan, gate, UI walk, ship, handoff, park, tidy, diet and setup skills; role agents; voice hook with `/alt-voice` |
| **alt-go** | comment discipline, goroutine ownership, Go proverbs with examples |
| **alt-oauth** | OAuth 2.1, OIDC and MCP authorization checklist by role; reviewer agent |
| **alt-mcp** | context7, Playwright and daisyUI docs MCP servers (no secrets) |
| **htmx** | htmx 4 guidance, debugging, extensions, upgrade from htmx 2 ([bigskysoftware/htmx](https://github.com/bigskysoftware/htmx) v4.0.2) |
| **go**, **cobra-viper**, **go-release**, **go-spec-reviewer** | from [spf13/go-skills](https://github.com/spf13/go-skills) |
| **daisyui** | from [saadeghi/daisyui](https://github.com/saadeghi/daisyui) |
| **antfu-frontend** | vue, pinia, vite, vitest, nuxt from [antfu/skills](https://github.com/antfu/skills) |

## Install

### Claude Code

```
/plugin marketplace add altalune-id/agent-plugins
/plugin install alt-workflow@altalune
```

Install any other plugin the same way. From the CLI, `claude plugin install <name>@altalune --scope project` enables it for one repo only.

### Codex

```
codex plugin marketplace add altalune-id/agent-plugins
```

Then `/plugins` in Codex and install the altalune plugins.

### OpenCode

```
npx skills add altalune-id/agent-plugins --skill '*' --agent opencode -g
```

### Gemini CLI

```
gemini skills install https://github.com/altalune-id/agent-plugins.git --path plugins/alt-workflow/skills
```

Repeat for `plugins/alt-go/skills`, `plugins/alt-oauth/skills` and `plugins/htmx/skills`.

### Cursor

Teams: an admin adds `https://github.com/altalune-id/agent-plugins` under Dashboard, Plugins & MCPs, Add Marketplace. Everyone else:

```
npx skills add altalune-id/agent-plugins --skill '*' --agent cursor -g
```

### Upstream skills outside Claude Code

The go, daisyui and antfu skills live in their own repos. Add them with the pin from `.claude-plugin/marketplace.json`:

```
npx skills add spf13/go-skills@<sha> --skill go --agent <codex|opencode|gemini-cli|cursor> -g
```

### Everything in one step (optional)

From a clone, `./install.sh` installs every plugin into Claude Code and into each of Codex, OpenCode, Gemini CLI and Cursor it finds on `PATH`, upstream skills and agents included. `./install.sh --dry-run` prints the commands first. Needs `jq`, and `go` for Codex or OpenCode agents.

- `--harness claude,codex` installs only into the named tools (`claude`, `codex`, `opencode`, `gemini`, `cursor`).
- `--scope project` installs into one repo instead of your user: Claude Code plugins at project scope, skills into `.agents/skills/`, agents into `.codex/agents/` and `.opencode/agents/`, each generated path added to the repo's `.git/info/exclude`.
- `--target <dir>` names that repo; the default is the working directory.

## What each tool supports

| | Skills | Agents | Hooks, `/alt-voice` | MCP servers |
| --- | :---: | :---: | :---: | :---: |
| Claude Code | ✓ | ✓ | ✓ | ✓ |
| Codex | ✓ | via `install.sh` | — | by hand |
| OpenCode | ✓ | via `install.sh` | — | by hand |
| Gemini CLI | ✓ | — | — | by hand |
| Cursor | ✓ | plugin route | — | by hand |

For MCP by hand, render the alt-mcp servers to a scratch file and merge them into your tool's config: `go run ./render mcp plugins/alt-mcp/.mcp.json --codex /tmp/mcp.toml` (or `--opencode /tmp/mcp.json`).

Install each skill once per tool: a plugin and the skills CLI together list every skill twice.

## Develop

```
make test              # render and installer tests
make validate          # claude plugin validate
make manifests         # regenerate Codex and Cursor manifests
make manifests-check   # CI drift check
```

Edit `.claude-plugin/marketplace.json` and each plugin's `.claude-plugin/plugin.json`; `make manifests` writes the rest.

## License

MIT. Vendored and upstream skills keep their own license.
