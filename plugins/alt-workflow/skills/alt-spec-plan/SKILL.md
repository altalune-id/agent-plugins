---
name: alt-spec-plan
description: Use after a design is approved and before implementation — turning an approved design into a reviewed spec and a phased plan with one PR per phase. Also use when asked to "write the plan" or "break this into tasks".
---

# Spec and plan

One plan per spec. The plan lists the decisions an implementer cannot make alone, nothing more.

## Spec

- Language skills first: for Go, invoke `go`, `cobra-viper`, `go-release`, `go-spec-reviewer` and the repo's convention skill.
- State behavior, interfaces at the boundaries, error cases, config keys, migrations, and the tests that prove it.
- Run `go-spec-reviewer` (or the stack's equivalent) on it. Fix findings or record why not. Tell the owner about any deviation from the approved design.

## Plan

File: `plans/<repo>/<topic>/plan.md`, first line `Status: in progress`.

Header:
- Goal (one sentence), architecture (two or three sentences), stack.
- Global constraints copied verbatim from the spec: versions, limits, naming and copy rules.
- Review focus: the five inputs or failure modes most likely to bite a user that no task's tests exercise. Each one gets a test in the task that owns the code.

Phases: one PR each, in dependency order. Tasks inside a phase are vertical slices: each ends with something a reviewer could accept or reject on its own.

Each task:
- Files to create or modify, exact paths.
- Interfaces it consumes and produces (exact names and types).
- The failing test first (name and assertions), the command to see it fail, the signature to implement, the command to see it pass.
- No function bodies unless the algorithm isn't determined by the signature and tests.

## Self-check before handing over

1. Every spec requirement maps to a task.
2. Names and types match across tasks.
3. No step says "handle edge cases" or "add tests" without naming them.
4. The plan is shorter than the code it describes.

## Hand over

Link the plan. The owner reviews it and says how to run it (lead alone, teammates, or subagents where the harness has no teammates, per `alt-lead` sizing). When its PRs merge: `Status: done`, then delete the plan folder.
