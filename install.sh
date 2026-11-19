#!/usr/bin/env bash
set -euo pipefail

MARKETPLACE_REPO="altalune-id/agent-plugins"
MARKETPLACE_NAME="altalune"
DRY_RUN=""

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    -h | --help)
      echo "usage: install.sh [--dry-run]"
      exit 0
      ;;
    *)
      echo "install.sh: unknown argument $arg" >&2
      exit 2
      ;;
  esac
done

cd "$(dirname "$0")"
CATALOG=".claude-plugin/marketplace.json"

run() {
  printf '+'
  printf ' %q' "$@"
  printf '\n'
  if [[ -z "$DRY_RUN" ]]; then
    "$@"
  fi
}

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "install.sh: $1 is required" >&2
    exit 1
  fi
}

need jq

marketplace_missing() {
  [[ -n "$DRY_RUN" ]] && return 0
  ! claude plugin marketplace list --json | jq -e --arg n "$MARKETPLACE_NAME" '.[] | select(.name == $n)' >/dev/null
}

if [[ -z "$DRY_RUN" ]]; then
  need claude
fi

if marketplace_missing; then
  run claude plugin marketplace add "$MARKETPLACE_REPO"
fi

while read -r plugin; do
  run claude plugin install "$plugin@$MARKETPLACE_NAME"
done < <(jq -r '.plugins[].name' "$CATALOG")

agents=()
for agent in codex opencode; do
  if command -v "$agent" >/dev/null 2>&1; then
    agents+=("$agent")
  fi
done

if [[ ${#agents[@]} -eq 0 ]]; then
  exit 0
fi

upstream_skills() {
  jq -r '
    .plugins[]
    | select(.source | type == "object")
    | .name as $plugin
    | .source as $s
    | ($s.repo // ($s.url | sub("^https://github.com/"; "") | sub("\\.git$"; ""))) as $repo
    | ($s.ref // $s.sha) as $pin
    | (if .skills then .skills[] | (if . == "." then $plugin else split("/") | last end) else $plugin end) as $skill
    | "\($repo)@\($pin) \($skill)"
  ' "$CATALOG"
}

while read -r source skill; do
  for agent in "${agents[@]}"; do
    run npx --yes skills add "$source" --skill "$skill" --agent "$agent" -g
  done
done < <(upstream_skills)

run mkdir -p "$HOME/.agents/skills"
for dir in plugins/*/skills/*/; do
  run cp -R "${dir%/}" "$HOME/.agents/skills/"
done

render_args=(agents plugins/alt-workflow/agents)
for agent in "${agents[@]}"; do
  case "$agent" in
    codex) render_args+=(--codex "$HOME/.codex/agents") ;;
    opencode) render_args+=(--opencode "$HOME/.config/opencode/agents") ;;
  esac
done
[[ -n "$DRY_RUN" ]] || need go
run go run ./render "${render_args[@]}"
