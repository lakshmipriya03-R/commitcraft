package cmd

import (
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

func newRecommitCmd() *cobra.Command {
	var (
		from    string
		dryRun  bool
		backup  bool
		noColor bool
	)

	cmd := &cobra.Command{
		Use:   "recommit",
		Short: "Interactively rewrite a range of commit messages",
		Long: `recommit walks you through every commit in a range and lets you
rewrite its message on the spot. When you're done, CommitCraft rebuilds
that slice of history using cherry-pick so the rest of your branch stays intact.

Example:
  # rewrite the last 5 commits interactively
  commitcraft recommit --from HEAD~5

  # dry-run to preview what would change
  commitcraft recommit --from HEAD~5 --dry-run`,

		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := viper.GetString("repo")
			verbose := viper.GetBool("verbose")

			gc, err := git.NewClient(repoPath, verbose)
			if err != nil {
				return err
			}

			// refuse to run against a dirty tree—filter-branch is unhappy and
			// so would the user be if we silently discarded their WIP
			clean, err := gc.IsClean()
			if err != nil {
				return err
			}
			if !clean {
				return fmt.Errorf("working tree has uncommitted changes—stash or commit them first")
			}

			if from == "" {
				return fmt.Errorf("--from is required (e.g. --from HEAD~5 or --from <commit-hash>)")
			}

			commits, err := gc.Log(from, "HEAD", 0)
			if err != nil {
				return err
			}
			if len(commits) == 0 {
				ui.Warn("no commits found in range %s..HEAD", from)
				return nil
			}

			ui.Header(fmt.Sprintf("recommit  —  %d commits to review", len(commits)))
			ui.Dim("Press Enter to keep the current message, or type a new one.")
			ui.Separator()

			type plan struct {
				commit     *git.Commit
				newMessage string
				changed    bool
			}

			plans := make([]plan, len(commits))

			for i, c := range commits {
				ui.CommitRow(c.Hash, c.Subject, c.Author, c.Timestamp)

				prompt := promptui.Prompt{
					Label:   fmt.Sprintf("  New message (%.7s)", c.Hash),
					Default: c.Message,
				}

				newMsg, promptErr := prompt.Run()
				if promptErr != nil {
					// user hit Ctrl+C—bail cleanly
					return fmt.Errorf("aborted by user")
				}

				plans[i] = plan{
					commit:     c,
					newMessage: newMsg,
					changed:    newMsg != c.Message,
				}

				if plans[i].changed {
					ui.DiffMessage(c.Message, newMsg)
				}
				ui.Separator()
			}

			// count how many are actually changing
			changeCount := 0
			for _, p := range plans {
				if p.changed {
					changeCount++
				}
			}

			if changeCount == 0 {
				ui.Info("no messages were changed—nothing to do")
				return nil
			}

			ui.Header(fmt.Sprintf("summary  —  %d of %d messages will change", changeCount, len(plans)))
			for _, p := range plans {
				if p.changed {
					ui.Dim("%.7s  %s", p.commit.Hash, p.newMessage)
				}
			}
			ui.Separator()

			if dryRun {
				ui.Warn("dry-run mode — no changes written")
				return nil
			}

			if !ui.Confirm(fmt.Sprintf("This will rewrite %d commits. Continue?", changeCount)) {
				ui.Info("aborted")
				return nil
			}

			branch, err := gc.CurrentBranch()
			if err != nil {
				return err
			}

			// create a safety backup branch before we start surgery
			if backup {
				backupName := fmt.Sprintf("commitcraft-backup-%d", time.Now().Unix())
				if err := gc.CreateTempBranch(backupName, "HEAD"); err != nil {
					return fmt.Errorf("could not create backup branch: %w", err)
				}
				ui.Info("backup branch created: %s", backupName)
				// switch back to where we were
				if err := gc.Checkout(branch); err != nil {
					return err
				}
			}

			// rewrite one commit at a time using filter-branch
			for _, p := range plans {
				if !p.changed {
					continue
				}
				ui.Info("rewriting %.7s ...", p.commit.Hash)
				if err := gc.RewriteCommitMessage(p.commit.Hash, p.newMessage); err != nil {
					ui.Error("failed to rewrite %.7s: %v", p.commit.Hash, err)
					return err
				}
			}

			gc.CleanupFilterBranch() //nolint:errcheck

			ui.Separator()
			ui.Success("done — %d commits rewritten on %s", changeCount, branch)
			ui.Dim("If you've already pushed, you'll need to force-push:")
			ui.Dim("  git push --force-with-lease origin %s", branch)

			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "start of commit range (e.g. HEAD~5 or a commit hash) [required]")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would change without writing anything")
	cmd.Flags().BoolVar(&backup, "backup", true, "create a backup branch before rewriting (recommended)")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable colored output")
	cmd.MarkFlagRequired("from") //nolint:errcheck

	return cmd
}
