package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lakshmipriya03-R/commitcraft/internal/git"
)

var rollbackBackup string

func newRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Restore the current branch from a backup branch",
		Long: `rollback resets the current branch back to a backup branch.
Use this when a rewrite goes wrong and you want to restore the previous history.

Example:
  commitcraft rollback --backup backup/master-20260624-170500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rollbackBackup == "" {
				return fmt.Errorf("--backup is required")
			}

			client, err := git.NewClient(viper.GetString("repo"), viper.GetBool("verbose"))
			if err != nil {
				return err
			}

			currentBranch, err := client.CurrentBranch()
			if err != nil {
				return err
			}

			ok, err := client.BranchExists(rollbackBackup)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("backup branch not found: %s", rollbackBackup)
			}

			fmt.Printf("\n  rollback\n")
			fmt.Printf("  ----------\n")
			fmt.Printf("  current branch : %s\n", currentBranch)
			fmt.Printf("  restore from   : %s\n\n", rollbackBackup)

			fmt.Printf("  This will reset '%s' to '%s'. Continue?\n", currentBranch, rollbackBackup)
			fmt.Printf("  Type 'yes' to continue: ")

			var answer string
			fmt.Scanln(&answer)
			if answer != "yes" {
				fmt.Println("  cancelled")
				return nil
			}

			if err := client.ResetBranchTo(currentBranch, rollbackBackup); err != nil {
				return err
			}

			fmt.Printf("  branch '%s' restored from '%s'\n", currentBranch, rollbackBackup)
			return nil
		},
	}

	cmd.Flags().StringVar(
		&rollbackBackup,
		"backup",
		"",
		"name of the backup branch to restore from",
	)

	return cmd
}
