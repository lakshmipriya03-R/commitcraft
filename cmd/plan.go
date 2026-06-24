package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

func newPlanCmd() *cobra.Command {
	var from string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Preview which commits would be rewritten",
		Long: `plan previews the commits that would be affected by a rewrite.
Use it before recommit/rewrite to understand what history will change.

Example:
  commitcraft plan --from HEAD~5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" {
				return fmt.Errorf("--from is required")
			}

			repoPath := viper.GetString("repo")
			verbose := viper.GetBool("verbose")

			gc, err := git.NewClient(repoPath, verbose)
			if err != nil {
				return err
			}

			commits, err := gc.Log(from, "HEAD", 0)
			if err != nil {
				return err
			}
			if len(commits) == 0 {
				ui.Warn("no commits found in the selected range")
				return nil
			}

			branch, _ := gc.CurrentBranch()
			ui.Header(fmt.Sprintf("rewrite plan  —  %s  (%d commits)", branch, len(commits)))
			ui.Separator()

			for _, c := range commits {
				ui.CommitRow(c.Hash, c.Subject, c.Author, c.Timestamp)
			}

			ui.Separator()
			ui.Dim("These commits would be rewritten if you run recommit/rewrite from this range.")
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "start of commit range (e.g. HEAD~5 or commit hash)")
	return cmd
}
