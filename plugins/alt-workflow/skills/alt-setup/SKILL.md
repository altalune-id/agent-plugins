---
name: alt-setup
description: Use when setting up the workspace on a new machine, after cloning the home repo, when repos are missing, or when the container runtime or test Postgres choice changes ("set up", "onboard", "bootstrap the workspace").
---

# Setup

The home repo's setup script does the work; this skill drives it and explains the choices. The script path is `paths.setup` in `workspace.json`.

## Before running

Ask the owner, one message, numbered, with a recommendation each:

1. Container runtime: `docker` or `podman`. Recommend the one already installed.
2. Test Postgres: `dev` (the dev stack's Postgres, shared, fast, no container per test) or `testcontainers` (one container per test binary, isolated, slower). Recommend `dev`.

## Run

```
SETUP_RUNTIME=<runtime> SETUP_PG=<dev|testcontainers> SETUP_NONINTERACTIVE=1 <setup>
```

It reads `workspace.json` and:
1. clones missing repos into their paths;
2. links `CLAUDE.md` to `AGENTS.md` where missing;
3. writes `.local/env` (container runtime, `TEST_PG_DSN` for `dev`) and creates the handoff folder;
4. installs the shared plugins (`install.sh` of the plugins repo) for Claude, Codex and opencode;
5. starts the dev stack and waits for Postgres.

Running it again changes nothing that is already in place.

## After

- Report what it cloned, linked and wrote.
- If the runtime socket is missing, the fix is a VM stop and start, then the dev stack up again.
- Never read `.env` files or secrets while doing this.
