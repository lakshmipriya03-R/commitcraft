package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

func newAuthorCmd() *cobra.Command {
	var (
		hash   string
		name   string
		email  string
		dryRun bool
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "author",
		Short: "Rewrite author name and email on one or more commits",
		Long: `author updates the GIT_AUTHOR and GIT_COMMITTER metadata for a commit.
Handy when you committed with the wrong email or before setting up your git config.

Example:
  # fix author on a specific commit
  commitcraft author --hash a3f1c9b --name "Alice Smith" --email "alice@example.com"

  # fix author on every commit in the entire repo (use carefully)
  commitcraft author --all --name "Alice Smith" --email "alice@example.com"`,

		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := viper.GetString("repo")
			verbose := viper.GetBool("verbose")

			gc, err := git.NewClient(repoPath, verbose)
			if err != nil {
				return err
			}

			if name == "" || email == "" {
				return fmt.Errorf("both --name and --email are required")
			}

			if !all && hash == "" {
				return fmt.Errorf("provide --hash or use --all to rewrite every commit")
			}

			if all {
				if !ui.Confirm("This will rewrite every commit in the repository. This is irreversible without a backup. Continue?") {
					ui.Info("aborted")
					return nil
				}

				if !dryRun {
					branch, err := gc.CurrentBranch()
					if err != nil {
						return err
					}

					backupBranch, err := gc.CreateTimestampedBackup(branch, branch)
					if err != nil {
						return fmt.Errorf("failed to create backup branch: %w", err)
					}
					ui.Warn("backup branch created: %s", backupBranch)

					if err := gc.RewriteAuthor("", name, email); err != nil {
						return err
					}
					gc.CleanupFilterBranch() //nolint:errcheck
				}

				ui.Success("all commits updated → %s <%s>", name, email)
				return nil
			}

			commit, err := gc.GetCommit(hash)
			if err != nil {
				return fmt.Errorf("cannot resolve %q: %w", hash, err)
			}

			ui.Header("author rewrite")
			ui.CommitRow(commit.Hash, commit.Subject, commit.Author, commit.Timestamp)
			ui.Dim("old: %s <%s>", commit.Author, commit.Email)
			ui.Dim("new: %s <%s>", name, email)
			ui.Separator()

			if dryRun {
				ui.Warn("dry-run mode — no changes written")
				return nil
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

			if !ui.Confirm(fmt.Sprintf("Rewrite author on %.7s?", commit.Hash)) {
				ui.Info("aborted")
				return nil
			}

			if err := gc.RewriteAuthor(commit.Hash, name, email); err != nil {
				return err
			}
			gc.CleanupFilterBranch() //nolint:errcheck

			ui.Success("author updated on %.7s", commit.Hash)
			return nil
		},
	}

	cmd.Flags().StringVar(&hash, "hash", "", "commit hash to update")
	cmd.Flags().StringVar(&name, "name", "", "new author name [required]")
	cmd.Flags().StringVar(&email, "email", "", "new author email [required]")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing")
	cmd.Flags().BoolVar(&all, "all", false, "apply to every commit in the repo")
	cmd.MarkFlagRequired("name")  //nolint:errcheck
	cmd.MarkFlagRequired("email") //nolint:errcheck

	return cmd
}
