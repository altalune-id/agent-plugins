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
agent_dirs="$(find plugins -mindepth 2 -maxdepth 2 -type d -name agents | wc -l | tr -d ' ')"
upstream="$(jq '[.plugins[] | select(.source | type == "object") | (.skills // [.name])[]] | length' .claude-plugin/marketplace.json)"

out="$(dry_run none)"
expect_line "$out" "+ claude plugin marketplace add altalune-id/agent-plugins"
expect_line "$out" "+ claude plugin install alt-workflow@altalune"
expect_count "$out" "^+ claude plugin install " "$plugins"
refuse_line "$out" "npx"
refuse_line "$out" "go run"
[[ "$(wc -l <<<"$out" | tr -d ' ')" == "$((plugins + 1))" ]] || fail "claude-only run printed extra lines"

out="$(dry_run all codex opencode gemini cursor-agent)"
expect_count "$out" "^+ claude plugin install " "$plugins"
for tool in codex opencode gemini-cli cursor; do
  expect_count "$out" "^+ npx --yes skills add .* --agent $tool -g -y$" "$((upstream + 1))"
  expect_line "$out" "+ npx --yes skills add altalune-id/agent-plugins --skill \\* --agent $tool -g -y"
done
expect_line "$out" "+ npx --yes skills add spf13/go-skills@383c2f8916313b7f1260a6d15d92a468e4ecc89e --skill go --agent codex -g -y"
expect_line "$out" "+ npx --yes skills add saadeghi/daisyui@v5.8.1 --skill daisyui --agent opencode -g -y"
expect_line "$out" "+ npx --yes skills add antfu/skills@4a6bb39e7b45e26402eb215393c51ccf5df1ca4c --skill pinia --agent gemini-cli -g -y"
refuse_line "$out" "cp -R"
expect_count "$out" "^+ go run ./render agents " "$agent_dirs"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex /home/tester/.codex/agents --opencode /home/tester/.config/opencode/agents"
expect_line "$out" "+ go run ./render agents plugins/alt-oauth/agents --codex /home/tester/.codex/agents --opencode /home/tester/.config/opencode/agents"

out="$(dry_run codex codex)"
expect_count "$out" "--agent codex" "$((upstream + 1))"
refuse_line "$out" "--agent opencode"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex /home/tester/.codex/agents"

out="$(dry_run gemini gemini)"
expect_count "$out" "--agent gemini-cli" "$((upstream + 1))"
refuse_line "$out" "go run"

if bash ./install.sh --bogus >/dev/null 2>&1; then
  fail "unknown argument accepted"
fi

mkbin() {
  local bin="$work/bin-$1"
  shift
  mkdir -p "$bin"
  ln -sf "$(command -v jq)" "$bin/jq"
  for tool in "$@"; do
    printf '#!/bin/sh\nexit 0\n' >"$bin/$tool"
    chmod +x "$bin/$tool"
  done
  echo "$bin"
}

fake_claude() {
  cat >"$1/claude" <<'SH'
#!/bin/sh
case "$*" in
  "plugin marketplace list --json") echo '[{"name":"altalune"}]' ;;
  "plugin list --json") printf '[{"id":"alt-workflow@altalune","scope":"user"},{"id":"htmx@altalune","scope":"project","projectPath":"%s"}]' "$FAKE_PROJECT" ;;
esac
exit 0
SH
  chmod +x "$1/claude"
}

expect_exit() {
  local want="$1" got=0
  shift
  "$@" >/dev/null 2>&1 || got=$?
  [[ "$got" == "$want" ]] || fail "exit $got, want $want: ${*: -3}"
}

repo="$PWD"
help="$(bash ./install.sh --help)"
for flag in --scope --target --harness --dry-run; do
  grep -q -- "$flag" <<<"$help" || fail "help lists no $flag"
done

bin="$(mkbin args)"
mkdir -p "$work/exists"
expect_exit 2 env PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --scope x
expect_exit 2 env PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --target "$work/exists"
expect_exit 2 env PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --scope project --target "$work/missing"
expect_exit 2 env PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --harness x
expect_exit 2 env PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --scope project --target=
expect_exit 2 env PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --harness ""

bin="$(mkbin filter codex opencode npx go)"
out="$(HOME=/home/tester PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --harness=codex)" || fail "harness filter: non-zero exit"
refuse_line "$out" "claude"
refuse_line "$out" "--agent opencode"
expect_line "$out" "+ npx --yes skills add altalune-id/agent-plugins --skill \\* --agent codex -g -y"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex /home/tester/.codex/agents"
HOME="$work/home" PATH="$bin:/usr/bin:/bin" bash ./install.sh --harness codex >/dev/null || fail "harness filter: needs claude without claude picked"

bin="$(mkbin project claude codex opencode gemini cursor-agent)"
mkdir -p "$work/target"
T="$(cd "$work/target" && pwd -P)"
out="$(HOME=/home/tester PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --scope project --target "$work/target")" || fail "project all: non-zero exit"
expect_line "$out" "+ claude plugin marketplace add altalune-id/agent-plugins"
expect_line "$out" "+ (cd $T) claude plugin install alt-workflow@altalune --scope project"
expect_count "$out" "^+ (cd $T) claude plugin install .* --scope project$" "$plugins"
expect_count "$out" "^+ (cd $T) npx --yes skills add .* -y$" "$((4 * (upstream + 1)))"
expect_count "$out" "npx" "$((4 * (upstream + 1)))"
expect_line "$out" "+ (cd $T) npx --yes skills add altalune-id/agent-plugins --skill \\* --agent gemini-cli -y"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex $T/.codex/agents --opencode $T/.opencode/agents"
refuse_line "$out" " -g"
refuse_line "$out" "/home/tester"

mkdir -p "$work/cwd"
C="$(cd "$work/cwd" && pwd -P)"
out="$(cd "$work/cwd" && HOME=/home/tester PATH="$bin:/usr/bin:/bin" bash "$repo/install.sh" --dry-run --scope project)" || fail "project default target: non-zero exit"
expect_line "$out" "+ (cd $C) claude plugin install alt-workflow@altalune --scope project"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex $C/.codex/agents --opencode $C/.opencode/agents"

bin="$(mkbin real codex opencode gemini cursor-agent npx go)"
fake_claude "$bin"
project_run() {
  FAKE_PROJECT="$(cd "$1" && pwd -P)" HOME="$work/home" PATH="$bin:/usr/bin:/bin" bash ./install.sh --scope project --target "$1"
}
patterns=(.agents/skills/alt-gate/ .agents/skills/go/ /skills-lock.json .codex/agents/alt-architect.toml .opencode/agents/alt-oauth-reviewer.md)
for dir in "$work/idem" "$work/my dir"; do
  git init -q "$dir"
  echo ".agents/skills/alt-gate/" >>"$dir/.git/info/exclude"
  out="$(project_run "$dir")" || fail "project run failed in $dir"
  refuse_line "$out" "install alt-workflow@altalune"
  refuse_line "$out" "install htmx@altalune"
  expect_line "$out" "+ ($(printf 'cd %q' "$(cd "$dir" && pwd -P)")) claude plugin install alt-go@altalune --scope project"
  for p in "${patterns[@]}"; do
    expect_count "$(cat "$dir/.git/info/exclude")" "^$p\$" 1
  done
  refuse_line "$out" "+ exclude .agents/skills/alt-gate/"
  expect_line "$out" "+ exclude .agents/skills/alt-design/"
  before="$(cat "$dir/.git/info/exclude")"
  out="$(project_run "$dir")" || fail "second project run failed in $dir"
  [[ "$(cat "$dir/.git/info/exclude")" == "$before" ]] || fail "second run changed exclude in $dir"
  refuse_line "$out" "+ exclude"
done
S="$(cd "$work/my dir" && pwd -P)"
out="$(HOME=/home/tester PATH="$bin:/usr/bin:/bin" bash ./install.sh --dry-run --scope project --target "$work/my dir")" || fail "project space target: non-zero exit"
expect_line "$out" "+ ($(printf 'cd %q' "$S")) npx --yes skills add altalune-id/agent-plugins --skill \\* --agent codex -y"
expect_line "$out" "+ go run ./render agents plugins/alt-workflow/agents --codex $(printf %q "$S/.codex/agents") --opencode $(printf %q "$S/.opencode/agents")"

if [[ "$failures" -gt 0 ]]; then
  echo "install_test: $failures failure(s)" >&2
  exit 1
fi
echo "install_test: ok"
