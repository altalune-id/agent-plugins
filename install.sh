#!/usr/bin/env bash
set -euo pipefail

MARKETPLACE_REPO="altalune-id/agent-plugins"
MARKETPLACE_NAME="altalune"
DRY_RUN=""
SCOPE="user"
TARGET=""
TARGET_SET=""
HARNESS=""

usage() {
  cat <<'EOF'
usage: install.sh [--scope user|project] [--target <dir>] [--harness <list>] [--dry-run]
  --scope    user (default) installs for your user; project installs into --target
  --target   repo for --scope project (default: the working dir)
  --harness  comma list of claude, codex, opencode, gemini, cursor (default: claude plus each tool on PATH)
  --dry-run  print the commands without running them
EOF
}

die() {
  echo "install.sh: $*" >&2
  exit 2
}

set_opt() {
  [[ -n "$2" ]] || die "$1 needs a value"
  case "$1" in
    --scope) SCOPE="$2" ;;
    --target) TARGET="$2" TARGET_SET=1 ;;
    --harness) HARNESS="$2" ;;
  esac
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    -h | --help)
      usage
      exit 0
      ;;
    --scope | --target | --harness)
      [[ $# -ge 2 ]] || die "$1 needs a value"
      set_opt "$1" "$2"
      shift
      ;;
    --scope=* | --target=* | --harness=*) set_opt "${1%%=*}" "${1#*=}" ;;
    *) die "unknown argument $1" ;;
  esac
  shift
done

case "$SCOPE" in
  user) [[ -z "$TARGET_SET" ]] || die "--target needs --scope project" ;;
  project) T="$(cd "${TARGET:-.}" 2>/dev/null && pwd -P)" || die "no such directory: $TARGET" ;;
  *) die "unknown scope $SCOPE" ;;
esac

CLAUDE=""
tools=()
if [[ -n "$HARNESS" ]]; then
  IFS=, read -ra picked <<<"$HARNESS"
  for h in "${picked[@]}"; do
    case "$h" in
      claude) CLAUDE=1 ;;
      codex | opencode | cursor) tools+=("$h") ;;
      gemini) tools+=(gemini-cli) ;;
      *) die "unknown harness $h" ;;
    esac
  done
else
  CLAUDE=1
  for pair in codex:codex opencode:opencode gemini:gemini-cli cursor-agent:cursor; do
    if command -v "${pair%%:*}" >/dev/null 2>&1; then
      tools+=("${pair#*:}")
    fi
  done
fi

cd "$(dirname "$0")"
CATALOG=".claude-plugin/marketplace.json"

run() {
  printf '+'
  printf ' %q' "$@"
  printf '\n'
  if [[ -z "$DRY_RUN" ]]; then
    "$@" </dev/null
  fi
}

run_in() {
  local dir="$1"
  shift
  printf '+ (cd %q)' "$dir"
  printf ' %q' "$@"
  printf '\n'
  if [[ -z "$DRY_RUN" ]]; then
    (cd "$dir" && "$@" </dev/null)
  fi
}

exclude() {
  local file="$T/.git/info/exclude"
  [[ -d "$T/.git" ]] || return 0
  if [[ -f "$file" ]] && grep -qxF -- "$1" "$file"; then
    return 0
  fi
  printf '+ exclude %s\n' "$1"
  [[ -z "$DRY_RUN" ]] || return 0
  mkdir -p "$T/.git/info"
  if [[ -s "$file" && -n "$(tail -c 1 "$file")" ]]; then
    printf '\n' >>"$file"
  fi
  printf '%s\n' "$1" >>"$file"
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

plugin_present() {
  [[ -n "$DRY_RUN" ]] && return 1
  jq -e --arg id "$1" --arg t "$T" 'any(.[]; .id == $id and (.scope == "user" or .projectPath == $t))' <<<"$installed" >/dev/null
}

if [[ -n "$CLAUDE" ]]; then
  if [[ -z "$DRY_RUN" ]]; then
    need claude
  fi

  if marketplace_missing; then
    run claude plugin marketplace add "$MARKETPLACE_REPO"
  fi

  if [[ "$SCOPE" == project && -z "$DRY_RUN" ]]; then
    installed="$(claude plugin list --json)"
  fi

  while read -r plugin; do
    if [[ "$SCOPE" == user ]]; then
      run claude plugin install "$plugin@$MARKETPLACE_NAME"
    elif ! plugin_present "$plugin@$MARKETPLACE_NAME"; then
      run_in "$T" claude plugin install "$plugin@$MARKETPLACE_NAME" --scope project
    fi
  done < <(jq -r '.plugins[].name' "$CATALOG")
fi

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

skills_add() {
  if [[ "$SCOPE" == project ]]; then
    run_in "$T" npx --yes skills add "$@" -y
  else
    run npx --yes skills add "$@" -g -y
  fi
}

while read -r source skill; do
  for tool in "${tools[@]}"; do
    skills_add "$source" --skill "$skill" --agent "$tool"
  done
done < <(upstream_skills)

for tool in "${tools[@]}"; do
  skills_add "$MARKETPLACE_REPO" --skill '*' --agent "$tool"
done

if [[ "$SCOPE" == project ]]; then
  exclude /skills-lock.json
  while read -r _ skill; do
    exclude ".agents/skills/$skill/"
  done < <(upstream_skills)
  for dir in plugins/*/skills/*/; do
    exclude ".agents/skills/$(basename "$dir")/"
  done
  codex_dir="$T/.codex/agents"
  opencode_dir="$T/.opencode/agents"
else
  codex_dir="$HOME/.codex/agents"
  opencode_dir="$HOME/.config/opencode/agents"
fi

targets=()
for tool in "${tools[@]}"; do
  case "$tool" in
    codex) targets+=(--codex "$codex_dir") ;;
    opencode) targets+=(--opencode "$opencode_dir") ;;
  esac
done
[[ ${#targets[@]} -gt 0 ]] || exit 0
[[ -n "$DRY_RUN" ]] || need go
for dir in plugins/*/agents; do
  run go run ./render agents "$dir" "${targets[@]}"
done

if [[ "$SCOPE" == project ]]; then
  for agent in plugins/*/agents/*.md; do
    name="$(basename "$agent" .md)"
    for tool in "${tools[@]}"; do
      case "$tool" in
        codex) exclude ".codex/agents/$name.toml" ;;
        opencode) exclude ".opencode/agents/$name.md" ;;
      esac
    done
  done
fi
