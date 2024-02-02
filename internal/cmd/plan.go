package cmd

import (
	"github.com/mach-composer/mach-composer-cli/internal/graph"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mach-composer/mach-composer-cli/internal/generator"
	"github.com/mach-composer/mach-composer-cli/internal/runner"
)

var planFlags struct {
	reuse      bool
	components []string
	lock       bool
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
	planCmd.Flags().BoolVarP(&planFlags.reuse, "reuse", "", false,
		"Suppress a terraform init for improved speed (not recommended for production usage)")
	planCmd.Flags().StringArrayVarP(&planFlags.components, "component", "c", nil, "")
	planCmd.Flags().BoolVarP(&planFlags.lock, "lock", "", true,
		"Acquire a lock on the state file before running terraform plan")
}

func planFunc(cmd *cobra.Command, _ []string) error {
	if len(planFlags.components) > 0 {
		log.Warn().Msgf("Components option not implemented")
	}

	cfg := loadConfig(cmd, true)
	defer cfg.Close()
	ctx := cmd.Context()

	dg, err := graph.ToDeploymentGraph(cfg, commonFlags.outputPath)
	if err != nil {
		return err
	}

	err = generator.Write(ctx, cfg, dg, nil)
	if err != nil {
		return err
	}

	b := runner.NewGraphRunner(commonFlags.workers)

	if err = checkReuse(ctx, dg, b, applyFlags.reuse); err != nil {
		return err
	}

	return b.TerraformPlan(ctx, dg, &runner.PlanOptions{
		Lock: planFlags.lock,
	})
}