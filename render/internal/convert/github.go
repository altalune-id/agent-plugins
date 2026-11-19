package convert

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitHub is a Fetcher that calls the GitHub API through the gh CLI.
type GitHub struct{}

// Tags lists the repository's tag names.
func (GitHub) Tags(ctx context.Context, owner, repo string) ([]string, error) {
	out, err := gh(ctx, "repos/"+owner+"/"+repo+"/tags", "--paginate", "--jq", ".[].name")
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// HeadSHA returns the newest commit sha on the repository's default branch.
func (GitHub) HeadSHA(ctx context.Context, owner, repo string) (string, error) {
	out, err := gh(ctx, "repos/"+owner+"/"+repo+"/commits/HEAD", "--jq", ".sha")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(out)
	if len(sha) != 40 {
		return "", fmt.Errorf("unexpected sha %q for %s/%s", sha, owner, repo)
	}
	return sha, nil
}

func gh(ctx context.Context, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "gh", append([]string{"api"}, args...)...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh api %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
