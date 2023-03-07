package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/labd/mach-composer/internal/generator"
	"github.com/labd/mach-composer/internal/runner"
)

var applyFlags struct {
	reuse       bool
	autoApprove bool
	destroy     bool
	components  []string
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the configuration.",
	PreRun: func(cmd *cobra.Command, args []string) {
		preprocessGenerateFlags()
	},

	Run: func(cmd *cobra.Command, args []string) {
		handleError(applyFunc(cmd.Context(), args))
	},
}

func init() {
	registerGenerateFlags(applyCmd)
	applyCmd.Flags().BoolVarP(&applyFlags.reuse, "reuse", "", false, "Supress a terraform init for improved speed (not recommended for production usage)")
	applyCmd.Flags().BoolVarP(&applyFlags.autoApprove, "auto-approve", "", false, "Supress a terraform init for improved speed (not recommended for production usage)")
	applyCmd.Flags().BoolVarP(&applyFlags.destroy, "destroy", "", false, "Destroy option is a convenient way to destroy all remote objects managed by this mach config")
	applyCmd.Flags().StringArrayVarP(&applyFlags.components, "component", "c", []string{}, "")
}

func applyFunc(ctx context.Context, args []string) error {
	cfg := loadConfig(ctx, true)
	defer cfg.Close()

	generateFlags.ValidateSite(cfg)

	// Note that we do this in multiple passes to minimize ending up with
	// half broken runs. We could in the future also run some parts in parallel

	paths, err := generator.WriteFiles(ctx, cfg, &generator.GenerateOptions{
		OutputPath: generateFlags.outputPath,
		Site:       generateFlags.siteName,
	})
	if err != nil {
		return err
	}

	return runner.TerraformApply(ctx, cfg, paths, &runner.ApplyOptions{
		Destroy:     applyFlags.destroy,
		Reuse:       applyFlags.reuse,
		AutoApprove: applyFlags.autoApprove,
		Site:        generateFlags.siteName,
		Components:  applyFlags.components,
	})
}
