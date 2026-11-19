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

tools=()
for pair in codex:codex opencode:opencode gemini:gemini-cli cursor-agent:cursor; do
  if command -v "${pair%%:*}" >/dev/null 2>&1; then
    tools+=("${pair#*:}")
  fi
done

if [[ ${#tools[@]} -eq 0 ]]; then
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
  for tool in "${tools[@]}"; do
    run npx --yes skills add "$source" --skill "$skill" --agent "$tool" -g -y
  done
done < <(upstream_skills)

for tool in "${tools[@]}"; do
  run npx --yes skills add "$MARKETPLACE_REPO" --skill '*' --agent "$tool" -g -y
done

targets=()
for tool in "${tools[@]}"; do
  case "$tool" in
    codex) targets+=(--codex "$HOME/.codex/agents") ;;
    opencode) targets+=(--opencode "$HOME/.config/opencode/agents") ;;
  esac
done
[[ ${#targets[@]} -gt 0 ]] || exit 0
[[ -n "$DRY_RUN" ]] || need go
for dir in plugins/*/agents; do
  run go run ./render agents "$dir" "${targets[@]}"
done
