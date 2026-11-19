---
name: alt-lead
description: Use when starting or continuing a piece of work as the lead — "do X", "continue <topic>", a kickoff brief, or any task that may need teammates or subagents. Sizes the work, decides who builds it, writes briefs, and keeps the lead's context small.
---

# Lead

The lead turns a request into finished, reviewed work. It decides, delegates and checks. It does not drift.

## Start

1. `continue <topic>`: read `<handoff_dir>/<topic>.md` (path from `workspace.json` `paths.handoff_dir`), then go on from its "Next" line. No handoff file: say so and ask what the topic is.
2. New work: restate the goal in one line. Check the backlog (`paths.backlog` in `workspace.json`), open issues and `plans/` for overlap; flag overlap or conflict and ask before building.
3. Read the `AGENTS.md` of every repo the work touches.

## Size the work

| Size | Who builds | Review |
|---|---|---|
| Up to about 3 files, or a fix round | the lead | one fresh reviewer agent plus the repo gates |
| A phase or a feature | teammates, one per lane | reviewer agents per phase, then the repo's merge reviewers |
| Any UI change | as above | plus `alt-ui-walk` |

Run work in parallel only when it is really independent. On a small machine, cap heavy lanes at two.

## Briefs

A teammate or subagent knows only its brief. Every brief carries:

- the goal and the definition of done (commands and expected output);
- the repo path and "read `AGENTS.md` first";
- the skills to invoke first, by name;
- hard limits: no commit or push unless the owner asked, no secrets, no destructive action;
- what is already ruled out;
- "report once at the end, under N words".

Teammates do not preload skills: the brief names them and says "invoke them with the Skill tool before coding".

## While work runs

- Keep teammates busy: when a review passes, the next task starts. Reviews run alongside the next task.
- Rule on anything that is not the owner's decision, and keep going.
- At each phase boundary, stop every idle teammate. Respawn with the brief when its lane has work again.
- The lead starts long-lived dev servers itself; stopping a teammate stops its servers.
- In a shared worktree, each builder stages only files from its own list.

## Decide well

- Smallest complete change that does what was asked. No polish, refactor or hardening nobody asked for. P2 and P3 findings go to the backlog.
- The code is the truth. A plan whose first line is `Status: in progress` is where the code is going.
- Leave notes, TODOs and plans you were not asked about alone; mention relevant ones at the end.
- Owner decisions only: public behavior, security, data, billing, anything hard to undo. Give options and a recommendation in chat.

## Finish

- Run `alt-gate` on the result.
- Report: what changed, proof (command output), what is left for the owner, and what you did not check.
- Long work pausing: write a handoff with `alt-handoff`.
