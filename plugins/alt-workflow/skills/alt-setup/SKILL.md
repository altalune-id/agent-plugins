---
name: alt-setup
description: Use when setting up a repo or workspace on a new machine, after cloning, when tools, repos, plugins or folder trust are missing, or when asked to set up, onboard, bootstrap or check a setup.
---

# Setup

The repo's own setup entry point does the work; this skill finds it, checks first, and explains each choice.

## 1. Find the entry point

First hit wins:

1. `paths.setup` in `workspace.json` at the repo root (`git rev-parse --show-toplevel`);
2. `bin/setup`;
3. `scripts/setup`;
4. a `setup` or `bootstrap` target in the `Makefile`;
5. the README's setup section.

Nothing found: say so and stop.

## 2. Check first

When `<entry> --help` lists `--check`, run `<entry> --check --json` (add `--json` only when the help lists it). Explain each `fail`: its id, detail and fix. Without `--check`, show the command and ask before running it.

## 3. Ask

One message, numbered: each fail with its recommended fix.

## 4. Apply on yes

Run the entry's non-interactive flags (`--yes` plus the answers), always with `--no-trust`. Never write folder trust from inside a session: the running agent may overwrite its own config. Print the trust command (`<entry> --trust --harness <h>`) for the user to run in a plain terminal.

## 5. Re-check

Run the step 2 check again. Report what passes, what fails and what was written. Say "restart your agent to apply" when `restart` is true, or when the entry wrote plugins, trust or settings.

## Rules

- Never read `.env` files or secrets.
- Never use `sudo`.
- Never edit the entry script.

## Per harness

Load `references/<harness>.md` for the harness running this skill: `claude-code`, `codex`, `opencode`, `gemini-cli` or `cursor`.
