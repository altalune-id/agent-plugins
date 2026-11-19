---
name: alt-comment-discipline
description: Use when writing, reviewing or trimming code comments in any language — godoc or docstrings, verbose or rationale comments, section headers, comment guidelines or style, a failing comment lint, or a request to clean up commenting in source, templ, SQL, YAML, Makefile, Dockerfile or .env.example files, even without an explicit ask for "comment cleanup".
---

# Comment discipline

**Default: no comment.** The code and its names are the source of truth. Comments rot silently:
nothing keeps them in step with the code. Why a change was made belongs in the commit message or
PR, where it is dated and reviewed, not in the file.

The test before typing `//` or `#`: _would deleting this lose anything the code and its names
do not already say?_

- No: do not write it.
- Yes, a hazard, unfinished work or an off-file reference: use a marker or the URL.
- Yes, a doc on an exported symbol: exactly one sentence.

## Keep

| Keep                                   | Shape                                                                                                            |
| -------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| doc on an exported / public symbol     | one sentence; in Go it starts with the identifier's name; covers func, method, type, const, var and package       |
| a marker                               | `TODO` `FIXME` `SECURITY` `NOTE` (also `XXX` `HACK` `BUG`)                                                       |
| an external reference                  | a URL, `RFC 1234`, a spec section, `GH-12`, `owner/repo#12`                                                      |
| a tool directive — not a comment       | `//go:build`, `// +build`, `//nolint`, `-- +goose Up`, `# syntax=`, `#!`, `# shellcheck` — no space after `//` |
| a config hint where the key is unclear | one line: `# 32 bytes, hex or base64`                                                                            |

Exported docs stay because tools read them (`go doc`, IDE hovers, pkg.go.dev), even when every
other comment goes.

## Delete

- Rationale, trade-offs, design or architecture notes.
- History: "added after the X incident", "was Y before".
- Comments that restate the code: `// increment counter`.
- Doc comments on private / unexported symbols, unless every line is a marker or URL.
- Decorative section headers: `// --- repositories ---`, `# ===== db =====`.
- Multi-line explanations of what a config block means.
- Blank `//` lines inside a doc block. Multi-line docs — keep one line. A marker that belongs
  with a doc joins the same line (see Examples); a second line fails a one-line-doc linter.

## Traps

- **A marker used as a loophole.** `// NOTE: we chose X because…` is still rationale. A
  `NOTE` or `SECURITY` line names a hazard the code does not show, in one line.
- **A wrong comment.** Worse than none, because it is trusted. Verify the claim before keeping
  or copying it from a neighboring file.
- **Directives that look like comments.** Build tags, goose markers, shebangs and linter
  pragmas are read by tools. Removing one breaks the build or the migration.
- **Generated files.** Never trim them by hand; fix the generator or leave them.
- **`// nolint` with a space** is prose, not a directive — linters only read `//nolint`.
- **Deprecation.** Go reads a paragraph starting `Deprecated:`. Under a one-line-doc rule, make
  it the whole doc: `// Deprecated: use Bar.`

## Examples

```go
// Before
// checkOwner enforces ownership here so no transport can skip it.
// Reported as absent, not forbidden, so callers learn nothing.
func (s *Service) checkOwner(...)

// After
// SECURITY: reports absent, not forbidden, so callers learn nothing about other rows.
func (s *Service) checkOwner(...)
```

The doc line goes because `checkOwner` is unexported; the `SECURITY` line stays because it
flags a hazard the code does not show.

```go
// Before
// BySlug resolves the public identifier.
//
// SECURITY: callers must check the row's owner against their grant.
func (s *Service) BySlug(...)

// After
// BySlug resolves the public identifier. SECURITY: callers must check the row's owner.
func (s *Service) BySlug(...)
```

```go
// Before
// Load resolves configuration: overrides > env > file > default.
//
// The mode is validated at startup so a typo fails fast.
func Load(opts Options) (Config, error)

// After
// Load resolves configuration: overrides > env > file > default.
func Load(opts Options) (Config, error)
```

## Sweeping existing code

1. Scope the files: `.go`, `.templ`, `.sql`, `.yaml`/`.yml`, `Dockerfile*`, `Makefile`,
   `.env.example`. Skip generated files (`// Code generated … DO NOT EDIT.`, `*_templ.go`,
   `*.min.js`, `*.min.css`), vendored trees (`vendor/`, `static/vendor/`), `bin/`, `go.mod`,
   `go.sum`, `.gitignore` and `.dockerignore`.
2. List violations with the repo's comment linter if it has one; it is the gate. Otherwise
   grep comment density per file and start from the densest:

   ```bash
   find . -name '*.go' -not -name '*_templ.go' | while read -r f; do
     echo "$(grep -cE '^\s*//' "$f") $f"
   done | sort -rn | head -20
   ```

3. Change comments only. Never touch code, struct fields, signatures or directives.
4. Build and test after each package (`go build ./...`, `go test ./...`) — a trimmed directive
   shows up there.
5. Report the violation or comment-line count before and after.

For a large sweep, give each package to a subagent with this file as its instructions.

## When not to apply

- Comments the user explicitly asks to keep.
- Tutorial or example code whose comments are the point.
- A public library whose godoc is its main documentation: a paragraph may be fair there, but
  aim for one sentence when one sentence is enough.
