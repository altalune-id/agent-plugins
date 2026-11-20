---
name: alt-tidy
description: Use after a piece of work lands, when asked to "tidy", "sweep" or "clean up", or when docs, memory, plans or checklists may be out of date. Finds what could confuse an agent and proposes fixes; also cleans machine leftovers.
---

# Tidy

Two jobs: keep written context true, and keep the machine clean. Nothing is changed or deleted until the owner says yes.

## Sweep (`/alt-tidy`)

1. **Find candidates** that could mislead an agent: TODOs in code, files under `plans/`, lines in `AGENTS.md` and `CLAUDE.md`, skills, memory files, backlog rows, open issues whose work merged.
2. **Prove each one.** A fresh subagent checks every candidate against the code and the git history. Plans are checked by their `Status:` line against what is built. Anything it cannot prove outdated stays.
3. **Propose** a list: archive, update or cut, each with its reason and evidence (file:line, commit, PR).
4. **On yes:** apply. Updates are fresh text: current truth in present tense, no history, no dates of change. History lives only in issues, specs, plans, ADRs and the git log. Archived files go to the archive folder (`paths.archive` in `workspace.json` when present; else ask the owner), which agents never read.

## Machine cleanup

Check first: no other heavy run active (the heavy-command lock, when the workspace has one), no other session using a local dev stack, if one exists.

Standing yes, report after:
- container volumes no container uses, when the container runtime is low on disk;
- `go clean -testcache` when free disk is under 6 GiB, plus an fstrim of the runtime's VM when it runs in one;
- a stale heavy-command lock, when the workspace has one (its owner process is dead), while a job waits.

Ask first:
- worktrees and merged local branches;
- containers and volumes your runs created that are still there (never any dev servers or local dev stack you didn't start);
- dangling images (`image prune`; never remove all images);
- a full Go build cache clean;
- scratch and Playwright output folders.

Report free disk before and after.
