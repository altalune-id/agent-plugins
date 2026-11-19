.PHONY: test validate render-check manifests manifests-check

test:
	go test -race ./render/...
	bash install_test.sh

validate:
	@if command -v claude >/dev/null 2>&1; then \
		claude plugin validate .; \
	else \
		for f in .claude-plugin/marketplace.json plugins/*/.claude-plugin/plugin.json; do \
			jq -e 'type == "object" and (.name | type == "string")' "$$f" >/dev/null || { echo "invalid: $$f"; exit 1; }; \
		done; \
		echo "json ok"; \
	fi

render-check:
	@tmp="$$(mktemp -d)"; trap 'rm -rf "$${tmp:?}"' EXIT; \
	go run ./render agents plugins/alt-workflow/agents --codex "$$tmp/codex" --opencode "$$tmp/opencode" >/dev/null; \
	want="$$(ls plugins/alt-workflow/agents/*.md | wc -l)"; \
	for d in codex opencode; do \
		got="$$(find "$$tmp/$$d" -type f -size +0 | wc -l)"; \
		[ "$$got" -eq "$$want" ] || { echo "$$d: $$got files, want $$want"; exit 1; }; \
	done; \
	echo "render ok: $$want agents"

manifests:
	go run ./render manifests .

manifests-check:
	go run ./render manifests . --check
