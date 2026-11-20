# OpenCode

| Topic | Fact |
| --- | --- |
| Skills | project skills in `.agents/skills/`; user skills in `~/.agents/skills/`; `npx skills ls` and `npx skills ls -g` list them |
| Agents | rendered by the plugins repo's `install.sh` into `.opencode/agents/` (project) or `~/.config/opencode/agents/` (user) |
| Folder trust | none: nothing to write |
| Start | start OpenCode at the workspace root so it finds the project skills |
| Restart | skills and agents are read at start: restart OpenCode after a change |
| Not supported | Claude hooks and `/alt-voice`; MCP servers are added to `opencode.json` by hand |
