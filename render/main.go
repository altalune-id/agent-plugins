// Command render converts Claude Code plugin pieces for Codex, OpenCode and Cursor and bumps upstream pins in the marketplace catalog.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"

	"github.com/altalune-id/agent-plugins/render/internal/convert"
)

const usage = `usage:
  render agents <src dir> [--codex <dir>] [--opencode <dir>]
  render mcp <src.json> [--codex <file>] [--opencode <file>]
  render bump <marketplace.json> [--write]
  render manifests <repo root> [--check]`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, convert.GitHub{}); err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, fetcher convert.Fetcher) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	switch args[0] {
	case "agents":
		return runAgents(args[1:], stdout, stderr)
	case "mcp":
		return runMCP(args[1:], stdout, stderr)
	case "bump":
		return runBump(ctx, args[1:], stdout, stderr, fetcher)
	case "manifests":
		return runManifests(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
}

func parse(fs *flag.FlagSet, args []string) (string, error) {
	var positional string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		positional, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if positional == "" && fs.NArg() > 0 {
		positional = fs.Arg(0)
	}
	if positional == "" {
		return "", fmt.Errorf("%s: missing source argument\n%s", fs.Name(), usage)
	}
	return positional, nil
}

func targets(fs *flag.FlagSet) (codex, opencode *string) {
	return fs.String("codex", "", "Codex output path"), fs.String("opencode", "", "OpenCode output path")
}

func runAgents(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("agents", flag.ContinueOnError)
	fs.SetOutput(stderr)
	codex, opencode := targets(fs)
	src, err := parse(fs, args)
	if err != nil {
		return err
	}
	if *codex == "" && *opencode == "" {
		return errors.New("agents: set --codex, --opencode or both")
	}
	written, err := convert.RenderAgents(src, *codex, *opencode)
	for _, w := range written {
		fmt.Fprintln(stdout, "wrote", w)
	}
	return err
}

func runMCP(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	codex, opencode := targets(fs)
	src, err := parse(fs, args)
	if err != nil {
		return err
	}
	if *codex == "" && *opencode == "" {
		return errors.New("mcp: set --codex, --opencode or both")
	}
	if err := convert.RenderMCP(src, *codex, *opencode); err != nil {
		return err
	}
	for _, w := range []string{*codex, *opencode} {
		if w != "" {
			fmt.Fprintln(stdout, "wrote", w)
		}
	}
	return nil
}

func runBump(ctx context.Context, args []string, stdout, stderr io.Writer, fetcher convert.Fetcher) error {
	fs := flag.NewFlagSet("bump", flag.ContinueOnError)
	fs.SetOutput(stderr)
	write := fs.Bool("write", false, "rewrite the catalog in place")
	path, err := parse(fs, args)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, changes, err := convert.Bump(ctx, data, fetcher)
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		fmt.Fprintln(stdout, "no changes")
		return nil
	}
	for _, c := range changes {
		fmt.Fprintln(stdout, c)
	}
	if !*write {
		return nil
	}
	return os.WriteFile(path, out, 0o644)
}

func runManifests(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("manifests", flag.ContinueOnError)
	fs.SetOutput(stderr)
	check := fs.Bool("check", false, "fail if a generated manifest is missing or stale, without writing")
	root, err := parse(fs, args)
	if err != nil {
		return err
	}
	files, err := convert.Manifests(root)
	if err != nil {
		return err
	}
	var stale []string
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if *check {
			if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, files[rel]) {
				stale = append(stale, rel)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, files[rel], 0o644); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "wrote", rel)
	}
	if len(stale) > 0 {
		return fmt.Errorf("manifests out of date, run render manifests: %s", strings.Join(stale, ", "))
	}
	return nil
}
