package runner

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/labd/mach-composer/internal/config"
)

type InitOptions struct {
	Site string
}

func TerraformInit(ctx context.Context, cfg *config.MachConfig, locations map[string]string, options *InitOptions) error {
	for i := range cfg.Sites {
		site := cfg.Sites[i]

		if options.Site != "" && site.Identifier != options.Site {
			continue
		}

		err := TerraformInitSite(ctx, cfg, &site, locations[site.Identifier], options)
		if err != nil {
			return err
		}
	}
	return nil
}

func TerraformInitSite(ctx context.Context, cfg *config.MachConfig, site *config.SiteConfig, path string, options *InitOptions) error {
	log.Debug().Msgf("Running terraform init for site %s", site.Identifier)

	return RunTerraform(ctx, path, "init")
}
