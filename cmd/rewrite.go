package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

func newRewriteCmd() *cobra.Command {
	var (
		hash    string
		message string
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "rewrite",
		Short: "Rewrite a single commit's message",
		Long: `rewrite updates the message of exactly one commit anywhere in history.
If the commit is HEAD, it uses git commit --amend. For older commits it
falls back to git filter-branch (which rewrites all descendant hashes).

Example:
  # rewrite the last commit
  commitcraft rewrite --hash HEAD --message "fix: correct off-by-one in pagination"

  # rewrite an older commit
  commitcraft rewrite --hash a3f1c9b --message "feat(auth): add JWT refresh token logic"`,

		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := viper.GetString("repo")
			verbose := viper.GetBool("verbose")

			gc, err := git.NewClient(repoPath, verbose)
			if err != nil {
				return err
			}

			if hash == "" {
				hash = "HEAD"
			}
			if message == "" {
				return fmt.Errorf("--message is required")
			}

			commit, err := gc.GetCommit(hash)
			if err != nil {
				return fmt.Errorf("cannot resolve %q: %w", hash, err)
			}

			ui.Header("rewrite")
			ui.CommitRow(commit.Hash, commit.Subject, commit.Author, commit.Timestamp)
			ui.DiffMessage(commit.Message, message)
			ui.Separator()

			if dryRun {
				ui.Warn("dry-run mode — no changes written")
				return nil
			}

			head, err := gc.HeadHash()
			if err != nil {
				return err
			}

			if commit.Hash == head {
				// simple case: amend HEAD
				ui.Info("amending HEAD")
				if err := gc.AmendCommit(message); err != nil {
					return err
				}
			} else {
				// older commit rewrite = destructive history rewrite
				clean, err := gc.IsClean()
				if err != nil {
					return err
				}
				if !clean {
					return fmt.Errorf("working tree has uncommitted changes—stash or commit them first")
				}

				branch, err := gc.CurrentBranch()
				if err != nil {
					return err
				}

				backupBranch, err := gc.CreateTimestampedBackup(branch, branch)
				if err != nil {
					return fmt.Errorf("failed to create backup branch: %w", err)
				}

				ui.Warn("backup branch created: %s", backupBranch)

				if !ui.Confirm(fmt.Sprintf(
					"Rewriting %.7s will change the hash of this commit and all descendants. Continue?",
					commit.Hash,
				)) {
					ui.Info("aborted")
					return nil
				}

				ui.Info("rewriting %.7s ...", commit.Hash)
				if err := gc.RewriteCommitMessage(commit.Hash, message); err != nil {
					return err
				}
				gc.CleanupFilterBranch() //nolint:errcheck
			}

			branch, _ := gc.CurrentBranch()
			ui.Success("commit rewritten on %s", branch)
			return nil
		},
	}

	cmd.Flags().StringVar(&hash, "hash", "HEAD", "commit hash (or ref) to rewrite")
	cmd.Flags().StringVarP(&message, "message", "m", "", "the new commit message [required]")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change without writing")
	cmd.MarkFlagRequired("message") //nolint:errcheck

	return cmd
}
