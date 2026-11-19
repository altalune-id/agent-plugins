package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

type fakeFetcher struct {
	tags  map[string][]string
	heads map[string]string
}

func (f fakeFetcher) Tags(_ context.Context, owner, repo string) ([]string, error) {
	tags, ok := f.tags[owner+"/"+repo]
	if !ok {
		return nil, fmt.Errorf("unexpected tags call for %s/%s", owner, repo)
	}
	return tags, nil
}

func (f fakeFetcher) HeadSHA(_ context.Context, owner, repo string) (string, error) {
	sha, ok := f.heads[owner+"/"+repo]
	if !ok {
		return "", fmt.Errorf("unexpected head call for %s/%s", owner, repo)
	}
	return sha, nil
}

func TestAgents(t *testing.T) {
	tests := []string{"full", "minimal", "alias"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("testdata", "agents", name)
			out := t.TempDir()
			var stdout, stderr bytes.Buffer
			args := []string{"agents", filepath.Join(dir, "in"), "--codex", filepath.Join(out, "codex"), "--opencode", filepath.Join(out, "opencode")}
			if err := run(t.Context(), args, &stdout, &stderr, fakeFetcher{}); err != nil {
				t.Fatalf("run: %v (stderr %q)", err, stderr.String())
			}
			compareDirs(t, out, filepath.Join(dir, "want"))
		})
	}
}

func TestMCP(t *testing.T) {
	tests := []string{"stdio", "http", "vars"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("testdata", "mcp", name)
			out := t.TempDir()
			var stdout, stderr bytes.Buffer
			args := []string{"mcp", filepath.Join(dir, "in.json"), "--codex", filepath.Join(out, "config.toml"), "--opencode", filepath.Join(out, "opencode.json")}
			if err := run(t.Context(), args, &stdout, &stderr, fakeFetcher{}); err != nil {
				t.Fatalf("run: %v (stderr %q)", err, stderr.String())
			}
			compareDirs(t, out, filepath.Join(dir, "want"))
		})
	}
}

func TestBump(t *testing.T) {
	fetcher := fakeFetcher{
		tags: map[string][]string{
			"acme/ui":   {"v5.10.0", "v5.9.3", "v6.0.0-beta.1", "latest", "v5.8.1"},
			"acme/same": {"v1.2.0", "v1.1.9"},
		},
		heads: map[string]string{
			"acme/skills": "1111111111111111111111111111111111111111",
			"acme/tools":  "2222222222222222222222222222222222222222",
		},
	}
	tests := []string{"tag", "sha", "local", "latest"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("testdata", "bump", name)
			out := t.TempDir()
			catalog := filepath.Join(out, "marketplace.json")
			copyFile(t, filepath.Join(dir, "in.json"), catalog)
			var stdout, stderr bytes.Buffer
			if err := run(t.Context(), []string{"bump", catalog, "--write"}, &stdout, &stderr, fetcher); err != nil {
				t.Fatalf("run: %v (stderr %q)", err, stderr.String())
			}
			if err := os.WriteFile(filepath.Join(out, "stdout.txt"), stdout.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			compareDirs(t, out, filepath.Join(dir, "want"))
		})
	}
}

func TestBumpWithoutWriteLeavesFile(t *testing.T) {
	fetcher := fakeFetcher{heads: map[string]string{
		"acme/skills": "1111111111111111111111111111111111111111",
		"acme/tools":  "2222222222222222222222222222222222222222",
	}}
	catalog := filepath.Join(t.TempDir(), "marketplace.json")
	copyFile(t, filepath.Join("testdata", "bump", "sha", "in.json"), catalog)
	var stdout, stderr bytes.Buffer
	if err := run(t.Context(), []string{"bump", catalog}, &stdout, &stderr, fetcher); err != nil {
		t.Fatalf("run: %v", err)
	}
	got, _ := os.ReadFile(catalog)
	want, _ := os.ReadFile(filepath.Join("testdata", "bump", "sha", "in.json"))
	if !bytes.Equal(got, want) {
		t.Fatal("catalog changed without --write")
	}
	if !strings.Contains(stdout.String(), "->") {
		t.Fatalf("stdout %q lists no change", stdout.String())
	}
}

var generatedManifests = []string{
	".agents/plugins/marketplace.json",
	".cursor-plugin/marketplace.json",
	"plugins/docs/.codex-plugin/plugin.json",
	"plugins/docs/.cursor-plugin/plugin.json",
	"plugins/tools/.codex-plugin/plugin.json",
	"plugins/tools/.cursor-plugin/plugin.json",
}

func TestManifests(t *testing.T) {
	golden := filepath.Join("testdata", "manifests", "repo")
	out := t.TempDir()
	if err := os.CopyFS(out, os.DirFS(golden)); err != nil {
		t.Fatal(err)
	}
	for _, f := range generatedManifests {
		if err := os.Remove(filepath.Join(out, f)); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	if err := run(t.Context(), []string{"manifests", out}, &stdout, &stderr, fakeFetcher{}); err != nil {
		t.Fatalf("run: %v (stderr %q)", err, stderr.String())
	}
	compareDirs(t, out, golden)
}

func TestManifestsCheck(t *testing.T) {
	golden := filepath.Join("testdata", "manifests", "repo")
	var stdout, stderr bytes.Buffer
	if err := run(t.Context(), []string{"manifests", golden, "--check"}, &stdout, &stderr, fakeFetcher{}); err != nil {
		t.Fatalf("check on golden repo: %v", err)
	}
	out := t.TempDir()
	if err := os.CopyFS(out, os.DirFS(golden)); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(out, "plugins", "docs", ".codex-plugin", "plugin.json")
	if err := os.WriteFile(stale, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run(t.Context(), []string{"manifests", out, "--check"}, &stdout, &stderr, fakeFetcher{})
	if err == nil || !strings.Contains(err.Error(), "plugins/docs/.codex-plugin/plugin.json") {
		t.Fatalf("check on stale repo: err = %v, want it to name the stale file", err)
	}
	got, _ := os.ReadFile(stale)
	if string(got) != "{}\n" {
		t.Fatal("--check rewrote a file")
	}
}

func TestRunErrors(t *testing.T) {
	src := filepath.Join("testdata", "agents", "minimal", "in")
	tests := []struct {
		name string
		args []string
	}{
		{"no command", nil},
		{"unknown command", []string{"nope"}},
		{"agents without targets", []string{"agents", src}},
		{"agents without source", []string{"agents", "--codex", t.TempDir()}},
		{"agents without frontmatter", []string{"agents", filepath.Join("testdata", "agents-bad"), "--codex", t.TempDir()}},
		{"agents empty dir", []string{"agents", t.TempDir(), "--codex", t.TempDir()}},
		{"mcp missing file", []string{"mcp", "testdata/none.json", "--codex", filepath.Join(t.TempDir(), "x.toml")}},
		{"bump missing file", []string{"bump", "testdata/none.json"}},
		{"manifests without source", []string{"manifests"}},
		{"manifests missing catalog", []string{"manifests", t.TempDir()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := run(t.Context(), tt.args, &stdout, &stderr, fakeFetcher{}); err == nil {
				t.Fatal("run succeeded, want error")
			}
		})
	}
}

func compareDirs(t *testing.T, got, want string) {
	t.Helper()
	if *update {
		if err := os.RemoveAll(want); err != nil {
			t.Fatal(err)
		}
		if err := os.CopyFS(want, os.DirFS(got)); err != nil {
			t.Fatal(err)
		}
		return
	}
	gotFiles := listFiles(t, got)
	wantFiles := listFiles(t, want)
	if strings.Join(gotFiles, "\n") != strings.Join(wantFiles, "\n") {
		t.Fatalf("files = %v, want %v", gotFiles, wantFiles)
	}
	for _, f := range wantFiles {
		g, _ := os.ReadFile(filepath.Join(got, f))
		w, _ := os.ReadFile(filepath.Join(want, f))
		if !bytes.Equal(g, w) {
			t.Errorf("%s differs\n--- got ---\n%s\n--- want ---\n%s", f, g, w)
		}
	}
}

func listFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, path)
		files = append(files, rel)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
