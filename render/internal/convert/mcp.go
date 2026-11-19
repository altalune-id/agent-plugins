package convert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
)

// MCPServer is one server entry of a Claude .mcp.json file.
type MCPServer struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func (s MCPServer) remote() bool { return s.URL != "" }

// ParseMCP reads a Claude .mcp.json document.
func ParseMCP(data []byte) (map[string]MCPServer, error) {
	var doc struct {
		MCPServers map[string]MCPServer `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.MCPServers) == 0 {
		return nil, errors.New("no mcpServers entries")
	}
	for name, s := range doc.MCPServers {
		if s.Command == "" && s.URL == "" {
			return nil, fmt.Errorf("server %q has neither command nor url", name)
		}
	}
	return doc.MCPServers, nil
}

// CodexMCP renders the servers as Codex config.toml mcp_servers tables.
func CodexMCP(servers map[string]MCPServer) []byte {
	var b strings.Builder
	for i, name := range slices.Sorted(maps.Keys(servers)) {
		s := servers[name]
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "[mcp_servers.%s]\n", tomlKey(name))
		if s.remote() {
			fmt.Fprintf(&b, "url = %s\n", tomlString(s.URL))
			if len(s.Headers) > 0 {
				fmt.Fprintf(&b, "http_headers = %s\n", tomlInlineTable(s.Headers))
			}
			continue
		}
		fmt.Fprintf(&b, "command = %s\n", tomlString(s.Command))
		if len(s.Args) > 0 {
			fmt.Fprintf(&b, "args = %s\n", tomlArray(s.Args))
		}
		if len(s.Env) > 0 {
			fmt.Fprintf(&b, "env = %s\n", tomlInlineTable(s.Env))
		}
	}
	return []byte(b.String())
}

type opencodeServer struct {
	Type        string            `json:"type"`
	Command     []string          `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// OpenCodeMCP renders the servers as an OpenCode config document with an mcp section.
func OpenCodeMCP(servers map[string]MCPServer) ([]byte, error) {
	out := map[string]opencodeServer{}
	for name, s := range servers {
		if s.remote() {
			out[name] = opencodeServer{Type: "remote", URL: s.URL, Headers: s.Headers}
			continue
		}
		out[name] = opencodeServer{Type: "local", Command: append([]string{s.Command}, s.Args...), Environment: s.Env}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(map[string]any{"mcp": out}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderMCP converts a Claude .mcp.json file into Codex TOML and OpenCode JSON files; an empty target is skipped.
func RenderMCP(src, codexFile, opencodeFile string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	servers, err := ParseMCP(data)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", src, err)
	}
	if codexFile != "" {
		if err := writeFile(codexFile, CodexMCP(servers)); err != nil {
			return err
		}
	}
	if opencodeFile == "" {
		return nil
	}
	doc, err := OpenCodeMCP(servers)
	if err != nil {
		return err
	}
	return writeFile(opencodeFile, doc)
}
