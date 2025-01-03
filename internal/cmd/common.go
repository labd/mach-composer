package cmd

import (
	"errors"
	"fmt"
	"github.com/mach-composer/mach-composer-cli/internal/cloud"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mach-composer/mach-composer-cli/internal/cli"
	"github.com/mach-composer/mach-composer-cli/internal/config"
)

type CommonFlags struct {
	configFile    string
	ignoreVersion bool
	outputPath    string
	varFile       string
	workers       int
}

var commonFlags CommonFlags

func registerCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&commonFlags.configFile, "file", "f", "main.yml", "YAML file to parse.")
	cmd.Flags().StringVarP(&commonFlags.varFile, "var-file", "", "", "Use a variable file to parse the configuration with.")
	cmd.Flags().BoolVarP(&commonFlags.ignoreVersion, "ignore-version", "", false, "Skip MACH composer version check")
	cmd.Flags().StringVarP(&commonFlags.outputPath, "output-path", "o", "deployments",
		"Outputs path to store the generated files.")
	cmd.Flags().IntVarP(&commonFlags.workers, "workers", "w", 1, "The number of workers to use")

	_ = cmd.RegisterFlagCompletionFunc("site", AutocompleteSiteName)
}

func preprocessCommonFlags(cmd *cobra.Command) {
	handleError(cmd.MarkFlagFilename("var-file", "yml", "yaml"))
	handleError(cmd.MarkFlagFilename("file", "yml", "yaml"))

	if _, err := os.Stat(commonFlags.configFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cli.PrintExitError(fmt.Sprintf("Config file %s does not exist", commonFlags.configFile))
		}
		cli.PrintExitError(err.Error())
	}
	if commonFlags.varFile != "" {
		if _, err := os.Stat(commonFlags.varFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cli.PrintExitError(fmt.Sprintf("Variable file %s does not exist", commonFlags.varFile))
			}
			log.Error().Msgf("error: %s\n", err.Error())
			os.Exit(1)
		}
	}
	if path.IsAbs(commonFlags.outputPath) == false {
		var err error
		value, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			commonFlags.outputPath = filepath.Join(value, commonFlags.outputPath)
		}
	}
}

func handleError(err error) {
	if err != nil {
		cli.PrintExitError("An error occurred:", err.Error())
	}
}

// loadConfig parses and validates the given config file path.
func loadConfig(cmd *cobra.Command, resolveVars bool) *config.MachConfig {
	opts := &config.ConfigOptions{
		NoResolveVars: !resolveVars,
		Validate:      true,
	}
	if commonFlags.varFile != "" {
		opts.VarFilenames = []string{commonFlags.varFile}
	}

	configFile, err := cmd.Flags().GetString("file")
	if err != nil {
		cli.PrintExitError("Missing config filename", err.Error())
	}

	cfg, err := config.Open(cmd.Context(), configFile, opts)
	if err != nil {
		cli.PrintExitError("An error occurred while loading the config file", err.Error())
	}

	if cfg.MachComposer.CloudEnabled() {
		if err := cloud.ResolveComponentsData(cmd.Context(), cfg); err != nil {
			cli.PrintExitError("An error occurred while fetching cloud component info", err.Error())
		}
	}

	return cfg
}
