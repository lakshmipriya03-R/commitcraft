package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

func newLintCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Scan recent commits for weak commit messages",
		Long: `lint checks recent commit messages and flags weak or low-quality subjects.

Examples of weak messages:
  fix
  update
  changes
  misc
  temp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gc, err := git.NewClient(viper.GetString("repo"), viper.GetBool("verbose"))
			if err != nil {
				return err
			}

			commits, err := gc.Log("", "HEAD", limit)
			if err != nil {
				return err
			}
			if len(commits) == 0 {
				ui.Warn("no commits found")
				return nil
			}

			ui.Header(fmt.Sprintf("commit lint  —  checking last %d commits", len(commits)))
			ui.Separator()

			badCount := 0
			for _, c := range commits {
				if isWeakMessage(c.Subject) {
					badCount++
					ui.Warn("weak commit message found")
					ui.CommitRow(c.Hash, c.Subject, c.Author, c.Timestamp)
					ui.Dim("suggestion: use a clearer message like 'feat: ...', 'fix: ...', or 'refactor: ...'")
					ui.Separator()
				}
			}

			if badCount == 0 {
				ui.Success("no weak commit messages found")
				return nil
			}

			ui.Warn("%d weak commit message(s) found", badCount)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "number", "n", 20, "number of recent commits to check")
	return cmd
}

func isWeakMessage(msg string) bool {
	msg = strings.TrimSpace(strings.ToLower(msg))

	weak := map[string]bool{
		"fix":     true,
		"update":  true,
		"changes": true,
		"misc":    true,
		"temp":    true,
		"test":    true,
		"work":    true,
		"wip":     true,
	}

	if weak[msg] {
		return true
	}

	if len(msg) <= 4 {
		return true
	}

	return false
}
