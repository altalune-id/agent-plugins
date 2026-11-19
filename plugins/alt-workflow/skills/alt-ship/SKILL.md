---
name: alt-ship
description: Use when work is ready to leave the machine — "do the PR", "push", "watch CI", "merge", or cleaning up after a merge. Branch, commit, PR, CI watch, bot replies, merge, cleanup.
---

# Ship

Commit, push and merge happen only when the owner asked for them in this conversation.

## Branch and commit

1. Branch from freshly pulled `main` (`git pull --ff-only`). One branch per PR.
2. Stage by explicit path from your own file list; check `git diff --cached --stat` against it. Never `git add -A`. Where `CLAUDE.md` is a link, stage `AGENTS.md`.
3. Sweep staged files for plan or milestone IDs and references to untracked files; write reasons in plain words.
4. One-line Conventional Commit, signed. Never `--no-verify` or `--no-gpg-sign`: fix what the hook reports.

## Push

- Push directly. A long pre-push hook can outlast the SSH idle timeout: `git -c core.sshCommand="ssh -o ServerAliveInterval=20 -o ServerAliveCountMax=60" push …`, then confirm with `git ls-remote`.

## PR

- Body follows the repo's PR template: summary, type, verification boxes ticked only for what you ran.
- Watch CI in the background (`gh pr checks <n> --watch`) and notify the owner on green or on a red run with its cause. Don't wait to be asked.
- Red run: read the failing job log, fix, push, watch again.
- Bot comments: fix the real ones; reply with the reason on false positives.

## Merge (on request)

- A green PR that follows the repo rules may merge with `gh pr merge --squash --admin` when the owner asks.
- After merge: `git switch main && git pull --ff-only`, delete the local branch, remove its worktree, prune (`git worktree prune`).
- Set the plan's `Status: done` when its last PR merged; delete the plan folder.

## Report

PR URL, CI state, what the owner still needs to do (env values as a copy-paste block, dashboard steps as numbered steps with exact values).
