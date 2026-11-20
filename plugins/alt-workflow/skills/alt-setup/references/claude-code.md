# Claude Code

| Topic | Fact |
| --- | --- |
| Skills | from plugins; also `.claude/skills/` (project) and `~/.claude/skills/` (user) |
| List plugins | `claude plugin list --json`: each entry has `id`, `scope` and, for project scope, `projectPath` |
| Install for one repo | `claude plugin install <plugin>@<marketplace> --scope project`, run in the repo; it writes `.claude/settings.json` |
| Folder trust | `~/.claude.json`, `projects["<absolute path>"].hasTrustDialogAccepted`; Claude Code asks on the first start in a folder |
| Restart | plugins, settings and trust are read at start: restart the session after a change |
| Supports | skills, agents, hooks, slash commands such as `/alt-voice`, MCP servers |
