# Gemini CLI

| Topic | Fact |
| --- | --- |
| Skills | project skills in `.agents/skills/`; `npx skills ls` and `npx skills ls -g` list them |
| Folder trust | `~/.gemini/trustedFolders.json`: absolute path to `TRUST_FOLDER`, `TRUST_PARENT` or `DO_NOT_TRUST`; a trusted ancestor covers its children; `DO_NOT_TRUST` means the user said no, so leave it |
| Change trust | `/permissions` inside Gemini CLI, or the entry's trust command in a plain terminal |
| Trust off | when `security.folderTrust.enabled` is false in `settings.json`, every folder is trusted and nothing needs writing |
| Restart | settings, trust and skills are read at start: restart Gemini CLI after a change |
| Not supported | agents, Claude hooks and `/alt-voice`; MCP servers are added to `settings.json` by hand |
