# agent-plugins

Shared skills, role agents and hooks for Go, Vue and htmx work. Claude Code is the source format; Codex and Cursor get generated plugin manifests that point at the same skill folders, and OpenCode and Gemini CLI install the skills through the [skills CLI](https://github.com/vercel-labs/skills). Upstream skills are referenced by git and pinned; a weekly job opens a PR when a newer tag or commit exists.

## Install

Pick one route per tool. Installing the same skills twice (a plugin and the skills CLI) gives every skill twice.

`./install.sh` (from a clone) installs everything into Claude Code, then into each of Codex, OpenCode, Gemini CLI and Cursor whose command (`codex`, `opencode`, `gemini`, `cursor-agent`) is on `PATH`, through the skills CLI. `--dry-run` prints the commands. It needs `jq`, and `go` when Codex or OpenCode is present.

| | Skills | Agents | Hooks and `/alt-voice` | MCP servers (alt-mcp) |
|---|---|---|---|---|
| Claude Code | all | yes | yes | yes |
| Codex | all | rendered TOML | no | by hand |
| OpenCode | all | rendered markdown | no | by hand |
| Gemini CLI | all | no | no | by hand |
| Cursor | all | plugin route only | no | by hand |

### Claude Code

```
/plugin marketplace add altalune-id/agent-plugins
/plugin install alt-workflow@altalune
```

Install each plugin you want the same way, or run `./install.sh` for all of them. You get every skill, the role agents, the voice hook with `/alt-voice`, and the alt-mcp servers.

### Codex

Plugin route, which updates with `codex plugin marketplace upgrade`:

```
codex plugin marketplace add altalune-id/agent-plugins
```

Then run `/plugins` in Codex and install htmx, alt-go, alt-workflow and alt-oauth from the altalune marketplace. Codex reads `.agents/plugins/marketplace.json` and each plugin's `.codex-plugin/plugin.json`, which point at the plugin's `skills/` folder. Add the upstream skills with the skills CLI (see [Upstream skills](#upstream-skills)).

Skills CLI route: `./install.sh` with `codex` on `PATH`. Skills land in `~/.agents/skills`.

Either route, then from a clone:

```
go run ./render agents plugins/alt-workflow/agents --codex ~/.codex/agents
go run ./render agents plugins/alt-oauth/agents --codex ~/.codex/agents
```

Codex can't:

- run plugin hooks in the CLI (the manifests turn them off), so there is no voice hook and no `/alt-voice`;
- give an agent its own MCP servers;
- install alt-mcp as a plugin. Render the servers to a scratch file and paste them into `~/.codex/config.toml`: `go run ./render mcp plugins/alt-mcp/.mcp.json --codex /tmp/alt-mcp.toml`. `render mcp` overwrites its target, so never point it at your real config.

### OpenCode

`./install.sh` with `opencode` on `PATH`, or by hand:

```
npx --yes skills add altalune-id/agent-plugins --skill '*' --agent opencode -g -y
go run ./render agents plugins/alt-workflow/agents --opencode ~/.config/opencode/agents
go run ./render agents plugins/alt-oauth/agents --opencode ~/.config/opencode/agents
```

Add the upstream skills with the skills CLI. OpenCode also reads skills from `.agents/skills`, `.claude/skills` and `.opencode/skills` in a project, and from `~/.agents/skills` and `~/.claude/skills`.

OpenCode can't run the Claude hooks (it needs a JavaScript plugin), so there is no voice hook and no `/alt-voice`. For MCP, run `go run ./render mcp plugins/alt-mcp/.mcp.json --opencode /tmp/alt-mcp.json` and merge its `mcp` section into `~/.config/opencode/opencode.json`.

### Gemini CLI

`./install.sh` with `gemini` on `PATH` (skills land in `~/.gemini/skills`), or with Gemini's own installer, one call per plugin:

```
gemini skills install https://github.com/altalune-id/agent-plugins.git --path plugins/alt-workflow/skills
```

Repeat with `plugins/alt-go/skills`, `plugins/alt-oauth/skills` and `plugins/htmx/skills`, and re-run to update. Add the upstream skills with the skills CLI.

Gemini CLI gets skills only. A Gemini extension needs a flat `skills/` folder at the repo root, so this repo ships no `gemini-extension.json`; agents, hooks and MCP servers are not installed. Add the alt-mcp servers to `~/.gemini/settings.json` by hand.

### Cursor

Plugin route, with skills and agents. On a Teams or Enterprise plan an admin opens Dashboard, Plugins & MCPs, Add Marketplace, Import from Repo, and enters `https://github.com/altalune-id/agent-plugins`; members then install from Customize. Without a team marketplace, copy a plugin from a clone into the local plugin folder and reload the window:

```
mkdir -p ~/.cursor/plugins/local
cp -R plugins/alt-workflow ~/.cursor/plugins/local/
```

Cursor reads `.cursor-plugin/marketplace.json` and each plugin's `.cursor-plugin/plugin.json`.

Skills CLI route, skills only: `./install.sh` with `cursor-agent` on `PATH`, or `npx --yes skills add altalune-id/agent-plugins --skill '*' --agent cursor -g -y`.

Cursor can't run the Claude hooks or `/alt-voice` (the manifests turn them off). Add the upstream skills with the skills CLI, and the alt-mcp servers to `~/.cursor/mcp.json` by hand.

### Upstream skills

go, cobra-viper, go-release, go-spec-reviewer, daisyui and antfu-frontend live in other repos. Claude Code installs them from the catalog. Every other tool gets them from the skills CLI at the pin in `.claude-plugin/marketplace.json`, which `./install.sh` reads for you:

```
npx --yes skills add <owner/repo>@<pin> --skill <name> --agent <codex|opencode|gemini-cli|cursor> -g -y
```

For example `npx --yes skills add spf13/go-skills@383c2f8916313b7f1260a6d15d92a468e4ecc89e --skill go --agent codex -g -y`. The antfu plugin gives the skills vue, pinia, vite, vitest and nuxt.

## Plugins

| Plugin | Source | What it gives |
|---|---|---|
| go, cobra-viper, go-release, go-spec-reviewer | spf13/go-skills (pinned sha) | Go idioms, CLIs, releases, spec review |
| daisyui | saadeghi/daisyui (pinned tag) | daisyUI 5 components and themes |
| antfu-frontend | antfu/skills (pinned sha) | Vue, Pinia, Vite, Vitest, Nuxt |
| htmx | bigskysoftware/htmx v4.0.2, vendored (0BSD) | htmx 4 guidance, debugging, extensions, upgrade |
| alt-go | this repo | comment discipline, goroutine ownership, Go proverbs |
| alt-workflow | this repo | lead, design, spec and plan, gate, UI walk, ship, handoff, park, tidy, diet; role agents; voice hook |
| alt-mcp | this repo | MCP servers with no secrets: context7, Playwright (isolated, output in `.playwright-mcp/`), daisyUI docs |
| alt-oauth | this repo | OAuth 2.1, OIDC and MCP authorization checklist by role; read-only reviewer agent |

## Develop

```
make test             # render tool and installer tests
make validate         # claude plugin validate
make manifests        # regenerate the Codex and Cursor manifests
make manifests-check  # fail if they are stale (CI runs it)
```

`render/` turns the Claude-format agents and MCP config into Codex and OpenCode files, writes the Codex and Cursor manifests, and bumps catalog pins. The manifests come from `.claude-plugin/marketplace.json` and each plugin's `.claude-plugin/plugin.json`: edit those, then run `make manifests`.

## License

MIT, except vendored skills, which keep their upstream license (see each plugin's LICENSE).
