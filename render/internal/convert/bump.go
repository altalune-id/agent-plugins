package convert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Fetcher looks up the tags and default-branch head of a GitHub repository.
type Fetcher interface {
	Tags(ctx context.Context, owner, repo string) ([]string, error)
	HeadSHA(ctx context.Context, owner, repo string) (string, error)
}

// Change is one pin moved by Bump.
type Change struct {
	Plugin string
	Field  string
	From   string
	To     string
}

func (c Change) String() string {
	return fmt.Sprintf("%s: %s %s -> %s", c.Plugin, c.Field, c.From, c.To)
}

var semverTag = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type pinnedSource struct {
	Source string `json:"source"`
	Repo   string `json:"repo"`
	URL    string `json:"url"`
	Ref    string `json:"ref"`
	SHA    string `json:"sha"`
}

// Bump moves semver refs to the newest semver tag and sha pins to the newest default-branch commit; the catalog's own formatting and key order are kept.
func Bump(ctx context.Context, catalog []byte, f Fetcher) ([]byte, []Change, error) {
	var doc struct {
		Plugins []json.RawMessage `json:"plugins"`
	}
	if err := json.Unmarshal(catalog, &doc); err != nil {
		return nil, nil, fmt.Errorf("parsing catalog: %w", err)
	}
	tags := map[string][]string{}
	heads := map[string]string{}
	out := catalog
	var changes []Change
	for _, raw := range doc.Plugins {
		var p struct {
			Name   string          `json:"name"`
			Source json.RawMessage `json:"source"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, nil, err
		}
		var src pinnedSource
		if bytes.HasPrefix(bytes.TrimSpace(p.Source), []byte(`"`)) || json.Unmarshal(p.Source, &src) != nil {
			continue
		}
		owner, repo, ok := githubRepo(src)
		if !ok {
			continue
		}
		key := owner + "/" + repo
		var field, from, to string
		switch {
		case src.SHA != "":
			if _, seen := heads[key]; !seen {
				sha, err := f.HeadSHA(ctx, owner, repo)
				if err != nil {
					return nil, nil, fmt.Errorf("%s: %w", p.Name, err)
				}
				heads[key] = sha
			}
			field, from, to = "sha", src.SHA, heads[key]
		case semverTag.MatchString(src.Ref):
			if _, seen := tags[key]; !seen {
				list, err := f.Tags(ctx, owner, repo)
				if err != nil {
					return nil, nil, fmt.Errorf("%s: %w", p.Name, err)
				}
				tags[key] = list
			}
			field, from, to = "ref", src.Ref, newestTag(src.Ref, tags[key])
		default:
			continue
		}
		if to == "" || to == from {
			continue
		}
		pin := regexp.MustCompile(`("` + field + `"\s*:\s*)"` + regexp.QuoteMeta(from) + `"`)
		updated := pin.ReplaceAll(raw, []byte(`${1}"`+to+`"`))
		out = bytes.Replace(out, raw, updated, 1)
		changes = append(changes, Change{Plugin: p.Name, Field: field, From: from, To: to})
	}
	return out, changes, nil
}

func githubRepo(src pinnedSource) (owner, repo string, ok bool) {
	path := src.Repo
	switch src.Source {
	case "github":
	case "url", "git-subdir":
		u, err := url.Parse(src.URL)
		if err != nil || u.Host != "github.com" {
			return "", "", false
		}
		path = strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	default:
		return "", "", false
	}
	owner, repo, ok = strings.Cut(path, "/")
	if !ok || owner == "" || repo == "" || strings.Contains(repo, "/") {
		return "", "", false
	}
	return owner, repo, true
}

func newestTag(current string, tags []string) string {
	best, bestKey := current, semverKey(current)
	for _, t := range tags {
		k := semverKey(t)
		if k != nil && compareKeys(k, bestKey) > 0 {
			best, bestKey = t, k
		}
	}
	return best
}

func semverKey(tag string) []int {
	m := semverTag.FindStringSubmatch(tag)
	if m == nil {
		return nil
	}
	key := make([]int, 3)
	for i := range key {
		key[i], _ = strconv.Atoi(m[i+1])
	}
	return key
}

func compareKeys(a, b []int) int {
	for i := range a {
		if a[i] != b[i] {
			return a[i] - b[i]
		}
	}
	return 0
}
