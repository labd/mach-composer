package generator

import (
	"context"
	"fmt"
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/utils"
	"strings"
)

// renderSite is responsible for generating the `site.tf` file. Therefore, it is
// the main entrypoint for generating the terraform file for each site.
func renderSite(ctx context.Context, cfg *config.MachConfig, site *config.SiteConfig, nestedComponents []config.SiteComponentConfig) (string,
	error) {
	result := []string{
		"# This file is auto-generated by MACH composer",
		fmt.Sprintf("# site: %s", site.Identifier),
	}

	// Render the terraform config
	val, err := renderSiteTerraformConfig(cfg, site)
	if err != nil {
		return "", fmt.Errorf("renderTerraformConfig: %w", err)
	}
	result = append(result, val)

	// Render all the file sources
	val, err = renderFileSources(cfg, site)
	if err != nil {
		return "", fmt.Errorf("failed to render file sources: %w", err)
	}
	result = append(result, val)

	// Render all the global resources
	val, err = renderSiteResources(cfg, site)
	if err != nil {
		return "", fmt.Errorf("failed to render resources: %w", err)
	}
	result = append(result, val)

	for _, component := range nestedComponents {
		if component.Deployment.Type != config.DeploymentSite {
			continue
		}
		val, err = renderComponentModule(ctx, cfg, site, &component)
		if err != nil {
			return "", fmt.Errorf("failed to render site component: %w", err)
		}
		result = append(result, val)
	}

	content := strings.Join(result, "\n")
	return content, nil
}

func renderSiteTerraformConfig(cfg *config.MachConfig, site *config.SiteConfig) (string, error) {
	var providers []string
	for _, plugin := range cfg.Plugins.All() {
		content, err := plugin.RenderTerraformProviders(site.Identifier)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render providers: %w", plugin.Name, err)
		}
		if content != "" {
			providers = append(providers, content)
		}
	}

	if !cfg.StateRepository.Has(site.Identifier) {
		return "", fmt.Errorf("state repository does not have a backend for site %s", site.Identifier)
	}
	backendConfig, err := cfg.StateRepository.Get(site.Identifier).Backend()
	if err != nil {
		return "", err
	}

	tpl, err := templates.ReadFile("templates/terraform.tmpl")
	if err != nil {
		return "", err
	}

	templateContext := struct {
		Providers     []string
		BackendConfig string
		IncludeSOPS   bool
	}{
		Providers:     providers,
		BackendConfig: backendConfig,
		IncludeSOPS:   cfg.Variables.HasEncrypted(site.Identifier),
	}
	return utils.RenderGoTemplate(string(tpl), templateContext)
}

func renderSiteResources(cfg *config.MachConfig, site *config.SiteConfig) (string, error) {
	var resources []string
	for _, plugin := range cfg.Plugins.All() {
		content, err := plugin.RenderTerraformResources(site.Identifier)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render resources: %w", plugin.Name, err)
		}

		if content != "" {
			resources = append(resources, content)
		}
	}

	tpl, err := templates.ReadFile("templates/resources.tmpl")
	if err != nil {
		return "", err
	}

	return utils.RenderGoTemplate(string(tpl), resources)
}