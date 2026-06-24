package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lakshmipriya03-R/commitcraft/internal/git"
)

// initTestRepo creates a fresh git repo in a temp dir, makes a few commits,
// and returns the Client pointed at it.
func initTestRepo(t *testing.T) (*git.Client, string) {
	t.Helper()

	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	// make three commits
	makeCommit := func(msg string) {
		f := filepath.Join(dir, msg+".txt")
		if err := os.WriteFile(f, []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{
			{"git", "add", "."},
			{"git", "commit", "-m", msg},
		} {
			c := exec.Command(args[0], args[1:]...)
			c.Dir = dir
			if out, err := c.CombinedOutput(); err != nil {
				t.Fatalf("commit %q: %s", msg, out)
			}
		}
	}

	makeCommit("first commit")
	makeCommit("second commit")
	makeCommit("third commit")

	gc, err := git.NewClient(dir, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return gc, dir
}

func TestNewClient_InvalidRepo(t *testing.T) {
	dir := t.TempDir() // just a dir, no git init
	_, err := git.NewClient(dir, false)
	if err == nil {
		t.Fatal("expected error for non-git directory, got nil")
	}
}

func TestLog_ReturnsCommits(t *testing.T) {
	gc, _ := initTestRepo(t)

	commits, err := gc.Log("", "HEAD", 10)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}
	if commits[0].Subject != "third commit" {
		t.Errorf("expected newest commit first, got %q", commits[0].Subject)
	}
}

func TestLog_WithLimit(t *testing.T) {
	gc, _ := initTestRepo(t)

	commits, err := gc.Log("", "HEAD", 2)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits with limit=2, got %d", len(commits))
	}
}

func TestGetCommit_HEAD(t *testing.T) {
	gc, _ := initTestRepo(t)

	c, err := gc.GetCommit("HEAD")
	if err != nil {
		t.Fatalf("GetCommit: %v", err)
	}
	if c.Subject != "third commit" {
		t.Errorf("unexpected subject: %q", c.Subject)
	}
	if c.Author != "Test User" {
		t.Errorf("unexpected author: %q", c.Author)
	}
}

func TestIsClean(t *testing.T) {
	gc, dir := initTestRepo(t)

	clean, err := gc.IsClean()
	if err != nil {
		t.Fatal(err)
	}
	if !clean {
		t.Error("expected clean repo, got dirty")
	}

	// write an untracked file
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("oops"), 0o644); err != nil {
		t.Fatal(err)
	}
	// stage it so git status --porcelain picks it up
	exec.Command("git", "-C", dir, "add", "dirty.txt").Run() //nolint:errcheck

	clean, err = gc.IsClean()
	if err != nil {
		t.Fatal(err)
	}
	if clean {
		t.Error("expected dirty repo, got clean")
	}
}

func TestCurrentBranch(t *testing.T) {
	gc, _ := initTestRepo(t)
	branch, err := gc.CurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	// git init creates either "main" or "master" depending on git config
	if branch != "main" && branch != "master" {
		t.Errorf("unexpected branch name: %q", branch)
	}
}

func TestCommit_TimestampParsed(t *testing.T) {
	gc, _ := initTestRepo(t)

	c, err := gc.GetCommit("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	// the timestamp should be recent (within the last minute)
	if time.Since(c.Timestamp) > time.Minute {
		t.Errorf("timestamp looks wrong: %v", c.Timestamp)
	}
}

func TestLog_CommitFields(t *testing.T) {
	gc, _ := initTestRepo(t)

	commits, err := gc.Log("", "HEAD", 1)
	if err != nil {
		t.Fatal(err)
	}
	c := commits[0]

	if len(c.Hash) != 40 {
		t.Errorf("hash should be 40 chars, got %d: %q", len(c.Hash), c.Hash)
	}
	if len(c.ShortHash) != 7 {
		t.Errorf("short hash should be 7 chars, got %d: %q", len(c.ShortHash), c.ShortHash)
	}
	if !strings.Contains(c.Email, "@") {
		t.Errorf("email looks wrong: %q", c.Email)
	}
}
