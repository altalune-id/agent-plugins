---
name: alt-diet
description: Use when sessions feel slow, distracted or heavy, when asked to "diet", "cut context", "why is the context so big", or after adding plugins, MCP servers, skills or memory. Measures what a fresh session loads and proposes cuts by size.
---

# Diet

Every byte a fresh session loads competes with the task. Measure, rank, cut what doesn't earn its place. Nothing changes until the owner says yes. Claude Code only: it measures what a Claude Code session loads. In another harness, say so and stop.

## Measure

1. Ask the owner to run `/context` in a fresh session and paste the breakdown. It is the only exact count of system prompt, tools, MCP tool names, memory and skills.
2. Run `measure.sh <project dir>` (next to this skill) for the static sources, per source in bytes and lines:
   - the `CLAUDE.md` / `AGENTS.md` chain (user, project, added directories when they load);
   - the memory index and memory files;
   - skill folders (project, user, enabled plugins) and the size of each `SKILL.md`;
   - MCP servers in `.mcp.json` and which are enabled;
   - hooks (user, project, plugins) and enabled plugins.

## Find

- Duplicates: the same skill from two sources, rules repeated across `AGENTS.md` files, memory that repeats a rule file.
- Oversized: a `SKILL.md` over 150 lines, a rules file over 200 lines, a memory index over 200 words.
- Always-on but rarely used: MCP servers with many tools a typical task doesn't need, plugins whose skills never trigger, connectors (mail, calendar, drive).
- Hook injections that repeat on every prompt or subagent.
- Stale: memory or rules about finished work, dates of change, past sessions (hand those to `alt-tidy`).
- Workflow stored in memory: memory is per person; shared workflow belongs in skills or `AGENTS.md`.

## Propose

A table ranked by bytes saved: source, size, cut (remove, disable by default, move, merge, shorten), what is lost, how to get it back. Recommend the top cuts.

## Apply (on yes)

- MCP: leave servers out of `enabledMcpjsonServers`; enable per task with `/mcp`.
- Plugins: `claude plugin disable <name>`.
- Skills: remove duplicates at the copy, never the source.
- Memory and rules: rewrite as fresh text, or move to the right home.

Re-measure and report before and after.
