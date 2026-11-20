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
| Polish, ideas, nits, cross-repo notes | the backlog, first home that applies: a card on `agentboard` when its MCP tools are available; else `paths.backlog` from `workspace.json` in the repo root when present (append a `\| Item \| Fix \| P \| Area \| Refs \|` row); else a GitHub issue in the current repo labeled `backlog` (create the label with `gh label create backlog` when missing) |
| Blocks a user, admin, deployer, auth or a journey, or leaves a feature half-done | not parked: fix it now |

Say which bucket each finding went to.

## Before writing

Search for a duplicate: `gh issue list -R <org>/<repo> --state all --search "<words>"` (it also finds `backlog` issues), and `card_list` on `agentboard` or grep the `paths.backlog` file. If one exists, add to it (a comment, a sharper card or row) instead of a new one.

## Backlog row

- With `agentboard`: `card_create` with title = item, body = fix, `priority`, `area`, `refs`, `repo` = owning repo.
- With `paths.backlog`: `| Item | Fix | P | Area | Refs |` in that file.
- With neither: a GitHub issue labeled `backlog`, title = item, body = fix, P, area and refs.
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
- Work that lands: close its issue with the PR link, or move its card to Done with `card_update` (without `agentboard`, delete its backlog row or close its `backlog` issue with the PR link).
- A cross-repo bug: one issue per repo, each linking the others.
