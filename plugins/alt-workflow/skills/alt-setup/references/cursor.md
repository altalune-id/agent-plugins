# Cursor

| Topic | Fact |
| --- | --- |
| Skills | project skills in `.agents/skills/`; user skills in `~/.cursor/skills/`; `npx skills ls` and `npx skills ls -g` list them |
| Teams | an admin can add the plugins repo as a team marketplace in the Cursor dashboard; then members install plugins there instead of the skills CLI |
| Folder trust | none: nothing to write |
| Restart | skills are read at start: restart Cursor after a change |
| Not supported | Claude hooks and `/alt-voice`; agents only through the plugin route; MCP servers are added by hand |
