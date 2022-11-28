package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/labd/mach-composer/cmd/mach-composer/cloudcmd"
)

var (
	name    = "mach-composer"
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	rootCmd = &cobra.Command{
		Use:   name,
		Short: "MACH composer is an orchestration tool for modern MACH ecosystems",
		Long: `MACH composer is a framework that you use to orchestrate and ` +
			`extend modern digital commerce & experience platforms, based on MACH ` +
			`technologies and cloud native services.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				panic(err)
			}
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}
			logrus.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp:       true,
				DisableLevelTruncation: true,
				DisableSorting:         true,
				PadLevelText:           true,
			})
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "Verbose output.")
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(cloudcmd.CloudCmd)
	rootCmd.AddCommand(componentsCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(sitesCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(terraformCmd)
	rootCmd.AddCommand(versionCmd)
}
