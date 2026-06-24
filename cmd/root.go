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
	Long: `
  ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗ ██████╗██████╗  █████╗ ███████╗████████╗
 ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝██╔════╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝
 ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║   ██║     ██████╔╝███████║█████╗     ██║   
 ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║   ██║     ██╔══██╗██╔══██║██╔══╝     ██║   
 ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║   ╚██████╗██║  ██║██║  ██║██║        ██║   
  ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝    ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   

  Craft cleaner Git history. One commit at a time.`,

	// don't show usage on errors—it's noisy and unhelpful
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
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("repo", rootCmd.PersistentFlags().Lookup("repo"))

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
