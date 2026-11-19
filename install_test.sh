#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0

fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

dry_run() {
  local name="$1"
  shift
  local bin="$work/$name"
  mkdir -p "$bin"
  ln -sf "$(command -v jq)" "$bin/jq"
  for tool in "$@"; do
    printf '#!/bin/sh\nexit 0\n' >"$bin/$tool"
    chmod +x "$bin/$tool"
  done
  HOME=/home/tester PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run
}

expect_line() {
  local out="$1" line="$2"
  grep -qxF -- "$line" <<<"$out" || fail "missing line: $line"
}

refuse_line() {
  local out="$1" pattern="$2"
  if grep -q -- "$pattern" <<<"$out"; then
    fail "unexpected line matching: $pattern"
  fi
}

expect_count() {
  local out="$1" pattern="$2" want="$3" got
  got="$(grep -c -- "$pattern" <<<"$out" || true)"
  [[ "$got" == "$want" ]] || fail "lines matching $pattern: got $got, want $want"
}

plugins="$(jq '.plugins | length' .claude-plugin/marketplace.json)"
local_skills="$(find plugins/*/skills -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"
upstream="$(jq '[.plugins[] | select(.source | type == "object") | (.skills // [.name])[]] | length' .claude-plugin/marketplace.json)"

out="$(dry_run none)"
expect_line "$out" "+ claude plugin marketplace add altalune-id/agent-plugins"
expect_line "$out" "+ claude plugin install alt-workflow@altalune"
expect_count "$out" "^+ claude plugin install " "$plugins"
refuse_line "$out" "npx"
refuse_line "$out" "go run"
[[ "$(wc -l <<<"$out" | tr -d ' ')" == "$((plugins + 1))" ]] || fail "claude-only run printed extra lines"

out="$(dry_run both codex opencode)"
expect_count "$out" "^+ claude plugin install " "$plugins"
expect_count "$out" "^+ npx --yes skills add .* --agent codex -g$" "$upstream"
expect_count "$out" "^+ npx --yes skills add .* --agent opencode -g$" "$upstream"
expect_line "$out" "+ npx --yes skills add spf13/go-skills@383c2f8916313b7f1260a6d15d92a468e4ecc89e --skill go --agent codex -g"
expect_line "$out" "+ npx --yes skills add saadeghi/daisyui@v5.8.1 --skill daisyui --agent opencode -g"
expect_line "$out" "+ npx --yes skills add antfu/skills@4a6bb39e7b45e26402eb215393c51ccf5df1ca4c --skill pinia --agent codex -g"
expect_line "$out" "+ mkdir -p /home/tester/.agents/skills"
expect_line "$out" "+ cp -R plugins/alt-workflow/skills/alt-gate /home/tester/.agents/skills/"
expect_count "$out" "^+ cp -R " "$local_skills"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex /home/tester/.codex/agents --opencode /home/tester/.config/opencode/agents"

out="$(dry_run codex codex)"
expect_count "$out" "--agent codex" "$upstream"
refuse_line "$out" "--agent opencode"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex /home/tester/.codex/agents"

if bash ./install.sh --bogus >/dev/null 2>&1; then
  fail "unknown argument accepted"
fi

if [[ "$failures" -gt 0 ]]; then
  echo "install_test: $failures failure(s)" >&2
  exit 1
fi
echo "install_test: ok"
