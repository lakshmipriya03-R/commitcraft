package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/lakshmipriya03-R/commitcraft/internal/git"
	"github.com/lakshmipriya03-R/commitcraft/internal/ui"
)

func newInspectCmd() *cobra.Command {
	var (
		limit  int
		from   string
		detail bool
	)

	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Display commit history with rich formatting",
		Long: `inspect shows recent commits in a clean, readable format—
useful before running recommit or rewrite to identify which hashes to target.

Example:
  # last 20 commits
  commitcraft inspect

  # last 50 commits with full message bodies
  commitcraft inspect -n 50 --detail

  # commits since a specific point
  commitcraft inspect --from HEAD~30`,

		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := viper.GetString("repo")
			verbose := viper.GetBool("verbose")

			gc, err := git.NewClient(repoPath, verbose)
			if err != nil {
				return err
			}

			commits, err := gc.Log(from, "HEAD", limit)
			if err != nil {
				return err
			}
			if len(commits) == 0 {
				ui.Warn("no commits found")
				return nil
			}

			branch, _ := gc.CurrentBranch()
			ui.Header(fmt.Sprintf("inspect  —  %s  (%d commits)", branch, len(commits)))
			ui.Separator()

			for _, c := range commits {
				ui.CommitRow(c.Hash, c.Subject, c.Author, c.Timestamp)

				if detail && c.Body != "" {
					// indent the body so it's clearly subordinate to the subject
					for _, line := range splitLines(c.Body) {
						ui.Dim("         %s", line)
					}
					ui.Separator()
				}
			}

			ui.Separator()
			ui.Dim("tip: run 'commitcraft recommit --from <hash>' to rewrite a range")

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "number", "n", 20, "number of commits to show")
	cmd.Flags().StringVar(&from, "from", "", "show commits after this ref (exclusive)")
	cmd.Flags().BoolVarP(&detail, "detail", "d", false, "show commit body text in addition to subject")

	return cmd
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
