package cmd

import (
	"context"
	"github.com/mach-composer/mach-composer-cli/internal/cli"
	"github.com/mach-composer/mach-composer-cli/internal/cmd/cloudcmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

var (
	RootCmd = &cobra.Command{
		Use:   "mach-composer",
		Short: "MACH composer is an orchestration tool for modern MACH ecosystems",
		Long: `MACH composer is a framework that you use to orchestrate and ` +
			`extend modern digital commerce & experience platforms, based on MACH ` +
			`technologies and cloud native services.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.Root().SilenceUsage = true
			cmd.Root().SilenceErrors = true

			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				cli.PrintExitError(err.Error())
			}

			quiet, err := cmd.Flags().GetBool("quiet")
			if err != nil {
				panic(err)
			}

			logger := log.Logger
			if verbose {
				logger = logger.Level(zerolog.TraceLevel)
			} else if quiet {
				logger = logger.Level(zerolog.ErrorLevel)
			} else {
				logger = logger.Level(zerolog.InfoLevel)
			}

			ctx := logger.WithContext(cmd.Context())
			logger = logger.Output(cli.NewConsoleWriter())

			o, err := cmd.Flags().GetString("output")
			if err != nil {
				cli.PrintExitError(err.Error())
			}

			output, err := cli.ConvertOutputType(o)
			if err != nil {
				cli.PrintExitError(err.Error())
			}

			ctx = cli.ContextWithOutput(ctx, output)

			//Github CI environment recognition
			if os.Getenv("GITHUB_ACTIONS") == "true" {
				ctx = cli.ContextWithGithubCI(ctx)
			}

			if output == cli.OutputTypeJSON {
				logger = logger.Output(os.Stdout)
			}
			log.Logger = logger
			// Register a signal handler to cancel the current context
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			ctx, cancel := context.WithCancel(ctx)

			go func() {
				select {
				case <-c:
					log.Info().Msg("Exiting...")
					cancel()
				case <-ctx.Done():
				}
			}()

			cmd.SetContext(ctx)
		},
	}
)

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output. This is equal to setting log levels to debug and higher")
	RootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet output. This is equal to setting log levels to error and higher")
	RootCmd.PersistentFlags().String("output", "console", "The output type. One of: console, json")
	RootCmd.AddCommand(applyCmd)
	RootCmd.AddCommand(cloudcmd.CloudCmd)
	RootCmd.AddCommand(componentsCmd)
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(planCmd)
	RootCmd.AddCommand(schemaCmd)
	RootCmd.AddCommand(showPlanCmd)
	RootCmd.AddCommand(sitesCmd)
	RootCmd.AddCommand(updateCmd)
	RootCmd.AddCommand(terraformCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(graphCmd)
	RootCmd.AddCommand(validateCmd)
}
