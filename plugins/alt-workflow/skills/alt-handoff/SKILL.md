---
name: alt-handoff
description: Use when pausing work for a fresh session, when the owner asks for a "fresh session prompt" or "handoff", or when context is getting long. Writes or reads a short state file instead of a pasted brief.
---

# Handoff

A handoff is a pointer file, not a copy of the conversation. Rules, skills and reviewers load from `AGENTS.md` and the plugins; the handoff never repeats them.

## Write

Path: `<handoff_dir>/<topic>.md`. `handoff_dir` is `paths.handoff_dir` from `workspace.json` in the repo root (`git rev-parse --show-toplevel`) when present, else `.local/now/`. With the default, add a `.local/` line to `.git/info/exclude` when it has none (local only, never committed, so the tree stays clean). Overwrite it; it holds the current state only.

```
# <topic>
Goal: <one line>
State: <what is done, as facts>
Next: <the one next action>
Open: <decisions waiting for the owner, numbered>
Pointers: <plan path, branch names, PR and issue links, worktree paths>
Running: <dev servers or stacks left running, with ports>
```

Under 40 lines. No secrets, no transcript excerpts, no history.

Then tell the owner: "start a fresh session and say `continue <topic>`".

## Read

`continue <topic>` (see `alt-lead`): read the file, check each pointer still exists (branch, PR state, worktree), say what changed since, then do "Next".

## Done

When the topic is finished, delete its handoff file.
