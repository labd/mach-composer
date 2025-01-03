package cmd

import (
	"github.com/mach-composer/mach-composer-cli/internal/batcher"
	"github.com/mach-composer/mach-composer-cli/internal/graph"
	"github.com/mach-composer/mach-composer-cli/internal/hash"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mach-composer/mach-composer-cli/internal/generator"
	"github.com/mach-composer/mach-composer-cli/internal/runner"
)

var planFlags struct {
	forceInit             bool
	site                  string
	components            []string
	lock                  bool
	ignoreChangeDetection bool
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan the configuration.",
	PreRun: func(cmd *cobra.Command, args []string) {
		preprocessCommonFlags(cmd)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return planFunc(cmd, args)
	},
}

func init() {
	registerCommonFlags(planCmd)
	planCmd.Flags().BoolVarP(&planFlags.forceInit, "force-init", "", false, "Force terraform initialization. By default mach-composer will reuse existing terraform resources")
	planCmd.Flags().StringVarP(&planFlags.site, "site", "s", "", "Site to parse. If not set parse all sites.")
	planCmd.Flags().StringArrayVarP(&planFlags.components, "component", "c", nil, "")
	planCmd.Flags().BoolVarP(&planFlags.lock, "lock", "", true, "Acquire a lock on the state file before running terraform plan")
	planCmd.Flags().BoolVarP(&planFlags.ignoreChangeDetection, "ignore-change-detection", "", false, "Ignore change detection to run even if the components are considered up to date")
}

func planFunc(cmd *cobra.Command, _ []string) error {
	if len(planFlags.components) > 0 {
		log.Warn().Msgf("Components option not implemented")
	}

	cfg := loadConfig(cmd, true)
	defer cfg.Close()
	ctx := cmd.Context()

	dg, err := graph.ToDeploymentGraph(cfg, commonFlags.outputPath, graph.WithTargetSiteName(planFlags.site))
	if err != nil {
		return err
	}

	err = generator.Write(ctx, cfg, dg, nil)
	if err != nil {
		return err
	}

	r := runner.NewGraphRunner(
		batcher.NaiveBatchFunc(),
		hash.Factory(cfg),
		commonFlags.workers,
	)

	return r.TerraformPlan(ctx, dg, &runner.PlanOptions{
		ForceInit:             planFlags.forceInit,
		Lock:                  planFlags.lock,
		IgnoreChangeDetection: planFlags.ignoreChangeDetection,
	})
}
