// Command render converts Claude Code plugin pieces for Codex and opencode and bumps upstream pins in the marketplace catalog.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/altalune-id/agent-plugins/render/internal/convert"
)

const usage = `usage:
  render agents <src dir> [--codex <dir>] [--opencode <dir>]
  render mcp <src.json> [--codex <file>] [--opencode <file>]
  render bump <marketplace.json> [--write]`

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
	return fs.String("codex", "", "Codex output path"), fs.String("opencode", "", "opencode output path")
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
