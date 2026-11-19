# agent-plugins

A Claude Code plugin marketplace with shared skills, role agents and hooks for Go, Vue and htmx work. Upstream skills are referenced by git and pinned; a weekly job opens a PR when a newer tag or commit exists.

## Install

Claude Code:

```
/plugin marketplace add altalune-id/agent-plugins
/plugin install alt-workflow@altalune
```

Claude Code, Codex and opencode in one step:

```
./install.sh            # --dry-run prints the commands
```

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
make test       # render tool and installer tests
make validate   # claude plugin validate
```

`render/` turns the Claude-format agents and MCP config into Codex and opencode files, and bumps catalog pins.

## License

MIT, except vendored skills, which keep their upstream license (see each plugin's LICENSE).
