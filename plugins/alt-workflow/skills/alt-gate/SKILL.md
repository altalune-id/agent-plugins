---
name: alt-gate
description: Use at the end of every task or plan phase, before saying anything is done, fixed or passing, and when acting on review comments. Covers tests first, debugging, the repo gates, a fresh reviewer, and evidence.
---

# Gate

"Done" means a fresh command output in this message proves it. A subagent's "success" is a claim, not proof.

## While building

- **Tests first** at the seams the plan agreed. Watch each new test fail for the right reason before writing the code. A test that can't fail is deleted.
- **Guard tests** (a test that pins a bug fix): break the fix and confirm the test fails, then restore.
- **Debugging:** first build a fast, deterministic loop that reproduces the failure. One hypothesis at a time, instrument the boundary between components. After three failed fixes, stop and question the design.

## Gates

1. The repo's own gates from its `AGENTS.md` (for example `make check`, `make lint`, `make comment-check`, UI tests).
2. Integration or container tests when a store, migration or boot path changed. Run them through the workspace's heavy-command wrapper when the repo says so. Count skipped tests: a skip is not a pass.
3. Read the full diff yourself.
4. A fresh reviewer agent gets the diff range, the task or spec text, and the repo rules, never the session history. For merge, also the repo's listed reviewers and a security review of what ships.
5. Fresh-text sweep on touched docs and skills: present tense, no history words, no dates of change.

## Acting on review

- Check each finding against the code before changing anything. Push back with evidence when it is wrong.
- Fix what blocks a user, an admin, a deployer, auth or a journey, or leaves a feature half-done. P2 and P3 go to the backlog, never a follow-up round.
- No performative agreement. State what you changed or why you didn't.

## Report

- Commands run and their result lines (pass, fail, skip counts).
- Reviewer verdict and what you did with each finding.
- What you did not check.
