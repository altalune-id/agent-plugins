package convert

import (
	"errors"
	"strconv"
	"strings"
)

type field struct {
	scalar string
	list   []string
	isList bool
}

func (f field) values() []string {
	if f.isList {
		return f.list
	}
	return splitTopLevel(f.scalar)
}

func splitFrontmatter(data string) (front, body string, err error) {
	data = strings.ReplaceAll(data, "\r\n", "\n")
	rest, ok := strings.CutPrefix(data, "---\n")
	if !ok {
		return "", "", errors.New("missing frontmatter opening ---")
	}
	if strings.HasPrefix(rest, "---\n") || rest == "---" {
		return "", strings.TrimPrefix(rest, "---\n"), nil
	}
	front, body, ok = strings.Cut(rest, "\n---\n")
	if ok {
		return front + "\n", body, nil
	}
	if f, found := strings.CutSuffix(rest, "\n---"); found {
		return f + "\n", "", nil
	}
	return "", "", errors.New("missing frontmatter closing ---")
}

func parseFrontmatter(front string) map[string]field {
	fields := map[string]field{}
	lines := strings.Split(front, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" || line[0] == ' ' || line[0] == '\t' || line[0] == '#' {
			continue
		}
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		rest = strings.TrimSpace(rest)
		var block []string
		for i+1 < len(lines) && (lines[i+1] == "" || lines[i+1][0] == ' ' || lines[i+1][0] == '\t') {
			i++
			block = append(block, lines[i])
		}
		switch {
		case rest == "":
			fields[key] = field{list: blockItems(block), isList: true}
		case rest[0] == '|' || rest[0] == '>':
			fields[key] = field{scalar: blockScalar(block, rest[0] == '>')}
		case rest[0] == '[':
			inner := strings.TrimSuffix(strings.TrimPrefix(rest, "["), "]")
			fields[key] = field{list: splitTopLevel(inner), isList: true}
		default:
			fields[key] = field{scalar: unquote(rest)}
		}
	}
	return fields
}

func blockItems(block []string) []string {
	var items []string
	indent := -1
	for _, l := range block {
		trimmed := strings.TrimLeft(l, " \t")
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}
		depth := len(l) - len(trimmed)
		if indent < 0 {
			indent = depth
		}
		if depth != indent {
			continue
		}
		if item, ok := strings.CutPrefix(trimmed, "- "); ok {
			items = append(items, unquote(strings.TrimSuffix(strings.TrimSpace(item), ":")))
			continue
		}
		if key, _, ok := strings.Cut(trimmed, ":"); ok {
			items = append(items, unquote(strings.TrimSpace(key)))
		}
	}
	return items
}

func blockScalar(block []string, fold bool) string {
	var parts []string
	for _, l := range block {
		parts = append(parts, strings.TrimSpace(l))
	}
	sep := "\n"
	if fold {
		sep = " "
	}
	return strings.TrimSpace(strings.Join(parts, sep))
}

func splitTopLevel(s string) []string {
	var out []string
	depth, start := 0, 0
	add := func(part string) {
		if part = unquote(strings.TrimSpace(part)); part != "" {
			out = append(out, part)
		}
	}
	for i, r := range s {
		switch r {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				add(s[start:i])
				start = i + 1
			}
		}
	}
	add(s[start:])
	return out
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'")
	}
	if before, _, ok := strings.Cut(s, " #"); ok {
		return strings.TrimSpace(before)
	}
	return s
}
