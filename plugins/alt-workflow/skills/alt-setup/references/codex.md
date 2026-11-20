# Codex

| Topic | Fact |
| --- | --- |
| Skills | project skills in `.agents/skills/`; `npx skills ls` lists project skills, `npx skills ls -g` user skills |
| Agents | rendered by the plugins repo's `install.sh` into `.codex/agents/` (project) or `~/.codex/agents/` (user) |
| Config home | `~/.codex/`, or `$CODEX_HOME` when set |
| Folder trust | `config.toml` in the config home: a `[projects."<absolute path>"]` table with `trust_level = "trusted"`; `untrusted` means the user said no, so leave it |
| Restart | config and skills are read at start: restart Codex after a change |
| Not supported | Claude hooks and `/alt-voice`; MCP servers are added to `config.toml` by hand |
