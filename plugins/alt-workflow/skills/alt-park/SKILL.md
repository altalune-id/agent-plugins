---
name: alt-park
description: Use when a finding, idea or follow-up should be recorded instead of done now — "park this", "backlog it", "raise an issue", review findings rated P2 or P3, or work found while doing something else.
---

# Park

Each item has one home: a backlog row or a GitHub issue, never both.

## Which home

| Item | Home |
|---|---|
| Work that will become a PR, or that users or reviewers see | GitHub issue in the owning repo |
| Polish, ideas, nits, cross-repo notes | a row in the backlog (`paths.backlog` in `workspace.json`) |
| Blocks a user, admin, deployer, auth or a journey, or leaves a feature half-done | not parked: fix it now |

Say which bucket each finding went to.

## Before writing

Search for a duplicate: `gh issue list -R <org>/<repo> --state all --search "<words>"` and grep the backlog. If one exists, add to it (a comment, or a sharper row) instead of a new one.

## Backlog row

`| Item | Fix | P | Area | Refs |` in the owning repo's table.
- P: `P1` blocks or half-done, `P2` real but not blocking, `P3` polish, `later` idea.
- Area: security, auth, api-mcp, web-ui, cli, ops, tests, docs, data, features.
- Refs: `repo#N`, `path:line`.
- Plain words: no milestone codes or plan IDs.

## Issue

- Title: what is wrong or wanted, in one line.
- Body: Problem, Fix (proposed), Done when (checkable). Cross-link related issues in sibling repos.
- Issues are public: show the owner the text first unless the owner asked to file it.

## Close the loop

- An item that becomes an issue leaves the backlog.
- Work that lands: close its issue with the PR link, or delete its backlog row.
- A cross-repo bug: one issue per repo, each linking the others.
