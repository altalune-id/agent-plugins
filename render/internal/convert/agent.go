package convert

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Agent is a Claude Code subagent definition read from markdown with YAML frontmatter.
type Agent struct {
	Name            string
	Description     string
	Tools           []string
	DisallowedTools []string
	Model           string
	Effort          string
	Skills          []string
	MCPServers      []string
	Body            string
}

var modelAliases = map[string]bool{"sonnet": true, "opus": true, "haiku": true, "fable": true, "inherit": true}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// ParseAgent reads a Claude agent file; fallbackName is used when the frontmatter has no name.
func ParseAgent(data []byte, fallbackName string) (Agent, error) {
	front, body, err := splitFrontmatter(string(data))
	if err != nil {
		return Agent{}, err
	}
	f := parseFrontmatter(front)
	a := Agent{
		Name:            f["name"].scalar,
		Description:     f["description"].scalar,
		Tools:           f["tools"].values(),
		DisallowedTools: f["disallowedTools"].values(),
		Model:           f["model"].scalar,
		Effort:          f["effort"].scalar,
		Skills:          f["skills"].values(),
		MCPServers:      f["mcpServers"].values(),
		Body:            body,
	}
	if a.Name == "" {
		a.Name = fallbackName
	}
	if !bareKey.MatchString(a.Name) {
		return Agent{}, fmt.Errorf("agent name %q is not a plain file name", a.Name)
	}
	return a, nil
}

// Codex renders the agent as a Codex custom agent TOML file.
func (a Agent) Codex() []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "name = %s\n", tomlString(a.Name))
	fmt.Fprintf(&b, "description = %s\n", tomlString(a.Description))
	switch {
	case a.Model == "":
	case modelAliases[a.Model]:
		fmt.Fprintf(&b, "# model alias %s not mapped\n", a.Model)
	default:
		fmt.Fprintf(&b, "model = %s\n", tomlString(a.Model))
	}
	if a.Effort != "" {
		effort := a.Effort
		if effort == "max" {
			effort = "xhigh"
		}
		fmt.Fprintf(&b, "model_reasoning_effort = %s\n", tomlString(effort))
	}
	for _, u := range []struct {
		name string
		set  []string
	}{{"tools", a.Tools}, {"disallowedTools", a.DisallowedTools}, {"skills", a.Skills}, {"mcpServers", a.MCPServers}} {
		if len(u.set) > 0 {
			fmt.Fprintf(&b, "# not supported in Codex: %s\n", u.name)
		}
	}
	fmt.Fprintf(&b, "developer_instructions = %s\n", tomlMultiline(strings.TrimSpace(a.Body)+"\n"))
	return []byte(b.String())
}

type permission struct {
	key      string
	action   string
	patterns [][2]string
}

// OpenCode renders the agent as an opencode subagent markdown file.
func (a Agent) OpenCode() []byte {
	var perms []*permission
	set := func(tool, action string) {
		key, pattern := opencodeTool(tool)
		var p *permission
		for _, existing := range perms {
			if existing.key == key {
				p = existing
			}
		}
		if p == nil {
			p = &permission{key: key}
			perms = append(perms, p)
		}
		if pattern == "" {
			p.action = action
			return
		}
		p.patterns = append(p.patterns, [2]string{pattern, action})
	}
	for _, t := range a.Tools {
		set(t, "allow")
	}
	for _, s := range a.MCPServers {
		set("mcp__"+s, "allow")
	}
	if len(a.Skills) > 0 {
		set("skill", "allow")
	}
	for _, t := range a.DisallowedTools {
		set(t, "deny")
	}

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "description: %s\n", strconv.Quote(a.Description))
	b.WriteString("mode: subagent\n")
	if len(perms) > 0 {
		b.WriteString("permission:\n")
	}
	for _, p := range perms {
		if len(p.patterns) == 0 {
			fmt.Fprintf(&b, "  %s: %s\n", yamlKey(p.key), p.action)
			continue
		}
		fmt.Fprintf(&b, "  %s:\n", yamlKey(p.key))
		if p.action != "" {
			fmt.Fprintf(&b, "    \"*\": %s\n", p.action)
		}
		for _, pat := range p.patterns {
			fmt.Fprintf(&b, "    %s: %s\n", strconv.Quote(pat[0]), pat[1])
		}
	}
	b.WriteString("---\n")
	b.WriteString(a.Body)
	return []byte(b.String())
}

func opencodeTool(tool string) (key, pattern string) {
	if name, rest, ok := strings.Cut(tool, "("); ok && strings.HasSuffix(rest, ")") {
		return strings.ToLower(name), strings.TrimSuffix(rest, ")")
	}
	if rest, ok := strings.CutPrefix(tool, "mcp__"); ok {
		server, name, found := strings.Cut(rest, "__")
		if !found {
			name = "*"
		}
		return strings.ToLower(server + "_" + name), ""
	}
	return strings.ToLower(tool), ""
}

func yamlKey(k string) string {
	if bareKey.MatchString(k) {
		return k
	}
	return strconv.Quote(k)
}

// RenderAgents converts every agent markdown file in src into the Codex and opencode directories; an empty target is skipped.
func RenderAgents(src, codexDir, opencodeDir string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(src, "*.md"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no agent markdown files in %s", src)
	}
	var written []string
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		a, err := ParseAgent(data, strings.TrimSuffix(filepath.Base(p), ".md"))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", p, err)
		}
		for _, out := range []struct {
			dir, ext string
			render   func() []byte
		}{{codexDir, ".toml", a.Codex}, {opencodeDir, ".md", a.OpenCode}} {
			if out.dir == "" {
				continue
			}
			target := filepath.Join(out.dir, a.Name+out.ext)
			if err := writeFile(target, out.render()); err != nil {
				return nil, err
			}
			written = append(written, target)
		}
	}
	return written, nil
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
