package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "commitcraft",
	Short: "Rewrite Git history with precision",
	Long:  "CommitCraft helps you rewrite Git history safely and cleanly.",
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.commitcraft.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().String("repo", ".", "path to the git repository")

	// bind flags to viper so they're accessible anywhere
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		fmt.Fprintln(os.Stderr, "failed to bind verbose flag:", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("repo", rootCmd.PersistentFlags().Lookup("repo")); err != nil {
		fmt.Fprintln(os.Stderr, "failed to bind repo flag:", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(
		newRecommitCmd(),
		newRewriteCmd(),
		newInspectCmd(),
		newAuthorCmd(),
		newTimestampCmd(),
		newRollbackCmd(),
		newPlanCmd(),
		newApplyCmd(),
		newLintCmd(),
	)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".commitcraft")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			color.HiBlack("using config: %s", viper.ConfigFileUsed())
		}
	}
}