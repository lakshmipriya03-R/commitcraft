package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client wraps git operations for a specific repository path.
type Client struct {
	RepoPath string
	Verbose  bool
}

// Commit holds the parsed data from a single git commit.
type Commit struct {
	Hash      string
	ShortHash string
	Author    string
	Email     string
	Timestamp time.Time
	Message   string
	Subject   string
	Body      string
}

// NewClient returns a Client pointed at the given repo path.
// It validates that path actually contains a git repository.
func NewClient(repoPath string, verbose bool) (*Client, error) {
	c := &Client{RepoPath: repoPath, Verbose: verbose}

	if _, err := c.run("rev-parse", "--git-dir"); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}
	return c, nil
}

// run executes a git subcommand and returns its stdout.
func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.RepoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if c.Verbose {
		fmt.Fprintf(os.Stderr, "  → git %s\n", strings.Join(args, " "))
	}

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Log returns parsed commits between two refs.
func (c *Client) Log(from, to string, limit int) ([]*Commit, error) {
	format := "%H%x00%h%x00%an%x00%ae%x00%at%x00%s%x00%b%x1e"

	args := []string{"log", "--format=" + format}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}

	rangeSpec := to
	if from != "" {
		rangeSpec = from + ".." + to
	}
	args = append(args, rangeSpec)

	raw, err := c.run(args...)
	if err != nil {
		return nil, err
	}

	if raw == "" {
		return nil, nil
	}

	var commits []*Commit
	records := strings.Split(raw, "\x1e")

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		parts := strings.Split(record, "\x00")
		if len(parts) < 6 {
			continue
		}

		ts, _ := parseUnixTimestamp(parts[4])
		body := ""
		if len(parts) > 6 {
			body = strings.TrimSpace(parts[6])
		}

		commits = append(commits, &Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Email:     parts[3],
			Timestamp: ts,
			Subject:   parts[5],
			Message:   buildFullMessage(parts[5], body),
			Body:      body,
		})
	}

	return commits, nil
}

// GetCommit fetches a single commit by ref.
func (c *Client) GetCommit(ref string) (*Commit, error) {
	commits, err := c.Log("", ref, 1)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("commit not found: %s", ref)
	}
	return commits[0], nil
}

// CurrentBranch returns the current branch name.
func (c *Client) CurrentBranch() (string, error) {
	return c.run("rev-parse", "--abbrev-ref", "HEAD")
}

// HeadHash returns HEAD hash.
func (c *Client) HeadHash() (string, error) {
	return c.run("rev-parse", "HEAD")
}

// IsClean returns true if working tree is clean.
func (c *Client) IsClean() (bool, error) {
	out, err := c.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

// AmendCommit rewrites HEAD commit message.
func (c *Client) AmendCommit(message string) error {
	_, err := c.run("commit", "--amend", "--no-edit", "--message", message)
	return err
}

// hasParent returns true if the given commit has a parent.
func (c *Client) hasParent(hash string) bool {
	_, err := c.run("rev-parse", hash+"^")
	return err == nil
}

// rewriteRange returns the correct filter-branch range.
// root commit -> hash..HEAD won't include root, so for root we use hash^? no.
// safest behavior:
// - if commit has parent => hash^..HEAD
// - if commit is root    => hash..HEAD is wrong, so use --all and filter inside script
func (c *Client) rewriteRange(hash string) string {
	if c.hasParent(hash) {
		return hash + "^..HEAD"
	}
	return "--all"
}

// RewriteCommitMessage rewrites one commit message anywhere in history.
func (c *Client) RewriteCommitMessage(hash, newMessage string) error {
	escaped := strings.ReplaceAll(newMessage, "'", "'\\''")

	filterScript := fmt.Sprintf(
		`if [ "$GIT_COMMIT" = "%s" ]; then printf '%%s' '%s'; else cat; fi`,
		hash, escaped,
	)

	args := []string{
		"filter-branch", "--force",
		"--msg-filter", filterScript,
	}

	rng := c.rewriteRange(hash)
	if rng == "--all" {
		args = append(args, "--", "--all")
	} else {
		args = append(args, rng)
	}

	_, err := c.run(args...)
	return err
}

// RewriteAuthor changes author name and email for a specific commit hash.
func (c *Client) RewriteAuthor(hash, name, email string) error {
	filterScript := fmt.Sprintf(`
if [ "$GIT_COMMIT" = "%s" ]; then
    export GIT_AUTHOR_NAME="%s"
    export GIT_AUTHOR_EMAIL="%s"
    export GIT_COMMITTER_NAME="%s"
    export GIT_COMMITTER_EMAIL="%s"
fi
`, hash, name, email, name, email)

	args := []string{
		"filter-branch", "--force",
		"--env-filter", filterScript,
	}

	rng := c.rewriteRange(hash)
	if rng == "--all" {
		args = append(args, "--", "--all")
	} else {
		args = append(args, rng)
	}

	_, err := c.run(args...)
	return err
}

// RewriteTimestamp updates author and committer date for a specific commit.
func (c *Client) RewriteTimestamp(hash string, ts time.Time) error {
	dateStr := ts.Format("2006-01-02T15:04:05-07:00")

	filterScript := fmt.Sprintf(`
if [ "$GIT_COMMIT" = "%s" ]; then
    export GIT_AUTHOR_DATE="%s"
    export GIT_COMMITTER_DATE="%s"
fi
`, hash, dateStr, dateStr)

	args := []string{
		"filter-branch", "--force",
		"--env-filter", filterScript,
	}

	rng := c.rewriteRange(hash)
	if rng == "--all" {
		args = append(args, "--", "--all")
	} else {
		args = append(args, rng)
	}

	_, err := c.run(args...)
	return err
}

// CreateTempBranch creates a new branch at the given ref.
func (c *Client) CreateTempBranch(name, ref string) error {
	_, err := c.run("checkout", "-b", name, ref)
	return err
}

// CreateBackupBranch creates a backup branch at the given ref.
func (c *Client) CreateBackupBranch(name, ref string) error {
	_, err := c.run("branch", name, ref)
	return err
}

// CreateTimestampedBackup creates a backup branch from the given ref and returns its name.
func (c *Client) CreateTimestampedBackup(branch, ref string) (string, error) {
	name := fmt.Sprintf("backup/%s-%s", branch, time.Now().Format("20060102-150405"))
	if err := c.CreateBackupBranch(name, ref); err != nil {
		return "", err
	}
	return name, nil
}

// BranchExists returns true if local branch exists.
func (c *Client) BranchExists(name string) (bool, error) {
	out, err := c.run("branch", "--list", name)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// ResetBranchTo resets branch to target.
func (c *Client) ResetBranchTo(branch, target string) error {
	if err := c.Checkout(branch); err != nil {
		return err
	}
	return c.ResetHard(target)
}

// DeleteBranch removes a local branch.
func (c *Client) DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := c.run("branch", flag, name)
	return err
}

// Checkout switches to ref.
func (c *Client) Checkout(ref string) error {
	_, err := c.run("checkout", ref)
	return err
}

// CherryPick applies a commit onto current HEAD.
func (c *Client) CherryPick(hash string) error {
	_, err := c.run("cherry-pick", hash)
	return err
}

// ResetHard hard-resets to ref.
func (c *Client) ResetHard(ref string) error {
	_, err := c.run("reset", "--hard", ref)
	return err
}

// CleanupFilterBranch removes backup refs left behind by filter-branch.
func (c *Client) CleanupFilterBranch() error {
	_, err := c.run("update-ref", "-d", "refs/original/refs/heads/main")
	if err != nil {
		c.run("update-ref", "-d", "refs/original/refs/heads/master") //nolint
	}
	return nil
}

// RemoteExists checks whether a named remote is configured.
func (c *Client) RemoteExists(name string) (bool, error) {
	out, err := c.run("remote")
	if err != nil {
		return false, err
	}
	for _, r := range strings.Split(out, "\n") {
		if strings.TrimSpace(r) == name {
			return true, nil
		}
	}
	return false, nil
}

// --- helpers ---

func parseUnixTimestamp(s string) (time.Time, error) {
	var unix int64
	_, err := fmt.Sscanf(s, "%d", &unix)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(unix, 0), nil
}

func buildFullMessage(subject, body string) string {
	if body == "" {
		return subject
	}
	return subject + "\n\n" + body
}
