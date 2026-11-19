#!/usr/bin/env bash
# Prints the static context sources a fresh Claude Code session loads for a project, with bytes and lines.
# Usage: measure.sh [project dir]. Env: CLAUDE_HOME (default ~/.claude).
set -uo pipefail

P=$(cd "${1:-.}" && pwd)
H=${CLAUDE_HOME:-$HOME/.claude}
key=$(printf '%s' "$P" | sed 's|[/.]|-|g')
total=0

row() {
  local label=$1 file=$2 b l
  [ -f "$file" ] || return 0
  b=$(wc -c < "$file" | tr -d ' ')
  l=$(wc -l < "$file" | tr -d ' ')
  total=$((total + b))
  printf '%-14s %8s %6s  %s\n' "$label" "$b" "$l" "$file"
}

printf '%-14s %8s %6s  %s\n' source bytes lines path
row rules "$H/CLAUDE.md"
row rules "$P/CLAUDE.md"
[ -L "$P/CLAUDE.md" ] || row rules "$P/AGENTS.md"
row memory "$H/projects/$key/memory/MEMORY.md"

for d in "$P/.claude/skills" "$H/skills"; do
  [ -d "$d" ] || continue
  while IFS= read -r f; do row skill "$f"; done < <(find -L "$d" -name SKILL.md 2>/dev/null | sort)
done

enabled=$(jq -r '.enabledPlugins // {} | to_entries[] | select(.value) | .key' "$H/settings.json" 2>/dev/null)
for id in $enabled; do
  name=${id%@*}; mkt=${id#*@}
  dir=$(find "$H/plugins/cache/$mkt/$name" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | tail -1)
  [ -n "$dir" ] || continue
  while IFS= read -r f; do row "plugin:$name" "$f"; done < <(find "$dir" -name SKILL.md 2>/dev/null | sort)
done

echo
echo "skill bytes above are full SKILL.md files; the session lists only name and description until a skill loads."
if [ -f "$P/.mcp.json" ]; then
  all=$(jq -r '.mcpServers // {} | keys[]' "$P/.mcp.json" | tr '\n' ' ')
  on=$(jq -r '.enabledMcpjsonServers // [] | .[]' "$P/.claude/settings.local.json" "$P/.claude/settings.json" 2>/dev/null | sort -u | tr '\n' ' ')
  echo "mcp servers:   $all"
  echo "mcp enabled:   ${on:-<all, none listed>}"
fi
echo "plugins on:    $(printf '%s ' $enabled)"
hooks=$(jq -r '.hooks // {} | keys[]' "$H/settings.json" "$P/.claude/settings.json" 2>/dev/null | sort | uniq -c | tr -s ' ' | tr '\n' ',')
echo "hook events:   ${hooks:-none in settings}"
echo "total bytes:   $total"
