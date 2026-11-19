---
description: "Reviews a change: correctness, tests and rules. Read-only."
mode: subagent
permission:
  read: allow
  grep: allow
  glob: allow
  bash:
    "*": allow
    "git push *": deny
  docs_search: allow
  "github_*": allow
  "docs_*": allow
  skill: allow
  write: deny
---

Load the skills first.

- Read every changed function in full.
- Report findings as a table.
