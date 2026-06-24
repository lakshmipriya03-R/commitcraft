package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/lakshmipriya03-R/commitcraft/internal/config"
	"github.com/lakshmipriya03-R/commitcraft/internal/git"
)

func newApplyCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply bulk rewrites from a YAML plan file",
		Long: `apply reads a YAML rewrite plan and performs commit rewrites.
A plan can update commit messages, author metadata, and timestamps.

Example:
  commitcraft apply --file rewrite.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			var plan config.RewritePlan
			if err := yaml.Unmarshal(data, &plan); err != nil {
				return fmt.Errorf("invalid YAML plan: %w", err)
			}

			if len(plan.Rewrites) == 0 {
				return fmt.Errorf("no rewrites found in plan")
			}

			gc, err := git.NewClient(viper.GetString("repo"), viper.GetBool("verbose"))
			if err != nil {
				return err
			}

			fmt.Printf("\napply plan\n")
			fmt.Printf("----------\n")
			fmt.Printf("Loaded %d rewrite entries from %s\n\n", len(plan.Rewrites), filePath)

			for _, item := range plan.Rewrites {
				if item.Hash == "" {
					continue
				}

				if item.Message != "" {
					fmt.Printf("Rewriting message for %s\n", item.Hash)
					if err := gc.RewriteCommitMessage(item.Hash, item.Message); err != nil {
						return err
					}
				}

				if item.AuthorName != "" || item.AuthorEmail != "" {
					fmt.Printf("Rewriting author for %s\n", item.Hash)
					if err := gc.RewriteAuthor(item.Hash, item.AuthorName, item.AuthorEmail); err != nil {
						return err
					}
				}

				if item.Timestamp != "" {
					fmt.Printf("Skipping timestamp for %s for now (we'll wire parsing next)\n", item.Hash)
				}
			}

			fmt.Printf("\nDone applying rewrite plan.\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "path to YAML rewrite plan")
	return cmd
}
