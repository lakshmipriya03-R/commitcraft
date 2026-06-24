package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

// accepted date input formats, tried in order
var dateLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02",
	time.RFC3339,
}

func newTimestampCmd() *cobra.Command {
	var (
		hash    string
		dateStr string
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "timestamp",
		Short: "Rewrite the date and time of a commit",
		Long: `timestamp changes the author and committer date on a specific commit.
Useful for fixing commits that were accidentally made at the wrong time,
or for organizing personal project history.

Date format: "2006-01-02 15:04:05" or "2006-01-02" (assumes midnight UTC)

Example:
  commitcraft timestamp --hash a3f1c9b --date "2024-03-15 09:30:00"
  commitcraft timestamp --hash HEAD --date "2024-03-15"`,

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
			if dateStr == "" {
				return fmt.Errorf("--date is required")
			}

			ts, err := parseDate(dateStr)
			if err != nil {
				return fmt.Errorf("could not parse date %q — try format: 2006-01-02 15:04:05", dateStr)
			}

			commit, err := gc.GetCommit(hash)
			if err != nil {
				return fmt.Errorf("cannot resolve %q: %w", hash, err)
			}

			ui.Header("timestamp rewrite")
			ui.CommitRow(commit.Hash, commit.Subject, commit.Author, commit.Timestamp)
			ui.Dim("old: %s", commit.Timestamp.Format("2006-01-02 15:04:05 -0700"))
			ui.Dim("new: %s", ts.Format("2006-01-02 15:04:05 -0700"))
			ui.Separator()

			if dryRun {
				ui.Warn("dry-run mode — no changes written")
				return nil
			}

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

			if !ui.Confirm(fmt.Sprintf("Rewrite timestamp on %.7s?", commit.Hash)) {
				ui.Info("aborted")
				return nil
			}

			if err := gc.RewriteTimestamp(commit.Hash, ts); err != nil {
				return err
			}
			gc.CleanupFilterBranch() //nolint:errcheck

			ui.Success("timestamp updated on %.7s", commit.Hash)
			return nil
		},
	}

	cmd.Flags().StringVar(&hash, "hash", "HEAD", "commit hash (or ref) to update")
	cmd.Flags().StringVar(&dateStr, "date", "", "new date/time (e.g. \"2024-03-15 09:30:00\") [required]")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would change without writing")
	cmd.MarkFlagRequired("date") //nolint:errcheck

	return cmd
}

func parseDate(s string) (time.Time, error) {
	for _, layout := range dateLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}
