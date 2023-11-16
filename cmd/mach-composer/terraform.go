package main

import (
	"github.com/mach-composer/mach-composer-cli/internal/dependency"
	"github.com/spf13/cobra"

	"github.com/mach-composer/mach-composer-cli/internal/runner"
)

var terraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Execute terraform commands directly",
	PreRun: func(cmd *cobra.Command, args []string) {
		preprocessGenerateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		handleError(terraformFunc(cmd, args))
	},
}

func init() {
	registerGenerateFlags(terraformCmd)
}

func terraformFunc(cmd *cobra.Command, args []string) error {
	cfg := loadConfig(cmd, true)
	defer cfg.Close()

	generateFlags.ValidateSite(cfg)

	dg, err := dependency.ToDeploymentGraph(cfg)
	if err != nil {
		return err
	}

	return runner.TerraformProxy(cmd.Context(), cfg, dg, &runner.ProxyOptions{
		Site:    generateFlags.siteName,
		Command: args,
	})
}
