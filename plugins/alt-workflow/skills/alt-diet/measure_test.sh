#!/usr/bin/env bash
# Tests measure.sh on a fixture project and Claude home with known file sizes.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
T=$(mktemp -d); trap 'rm -rf "$T"' EXIT
fail=0
check() { if eval "$2"; then echo "ok   $1"; else echo "FAIL $1"; fail=1; fi; }

P=$T/proj; H=$T/home
mkdir -p "$P/.claude/skills/a" "$H/skills/b" "$H/plugins/cache/mk/pl/1.0/skills/c"
key=$(printf '%s' "$P" | sed 's|[/.]|-|g'); mkdir -p "$H/projects/$key/memory"
printf '1234567890' > "$P/AGENTS.md"; ln -s AGENTS.md "$P/CLAUDE.md"
printf '12345' > "$H/projects/$key/memory/MEMORY.md"
printf 'abc\n' > "$P/.claude/skills/a/SKILL.md"
printf 'abcd\n' > "$H/skills/b/SKILL.md"
printf 'ab\n' > "$H/plugins/cache/mk/pl/1.0/skills/c/SKILL.md"
echo '{"enabledPlugins":{"pl@mk":true,"off@mk":false},"hooks":{"SessionStart":[]}}' > "$H/settings.json"
echo '{"mcpServers":{"x":{},"y":{}}}' > "$P/.mcp.json"
echo '{"enabledMcpjsonServers":["x"]}' > "$P/.claude/settings.local.json"

out=$(CLAUDE_HOME=$H bash "$HERE/measure.sh" "$P")
# CLAUDE.md symlink (10) + memory (5) + skills 4 + 5 + 3
check "total bytes" 'printf "%s" "$out" | grep -q "total bytes:   27"'
check "symlinked AGENTS.md counted once" '[ "$(printf "%s" "$out" | grep -c "^rules")" -eq 1 ]'
check "plugin skill listed" 'printf "%s" "$out" | grep -q "^plugin:pl"'
check "disabled plugin ignored" '! printf "%s" "$out" | grep -q "plugin:off"'
check "mcp enabled list" 'printf "%s" "$out" | grep -q "mcp enabled:   x"'
check "hook events" 'printf "%s" "$out" | grep -q "SessionStart"'
exit $fail
