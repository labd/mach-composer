package generator

import (
	"context"
	"fmt"
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/graph"
	"github.com/mach-composer/mach-composer-cli/internal/utils"
	"runtime"
	"slices"
	"strings"
)

type componentContext struct {
	ComponentName       string
	ComponentVersion    string
	ComponentHash       string
	ComponentVariables  string
	ComponentSecrets    string
	SiteName            string
	Environment         string
	Source              string
	PluginResources     []string
	PluginProviders     []string
	PluginDependsOn     []string
	PluginVariables     []string
	HasCloudIntegration bool
}

func renderSiteComponent(ctx context.Context, cfg *config.MachConfig, n *graph.SiteComponent) (string, error) {
	result := []string{
		"# This file is auto-generated by MACH composer",
		fmt.Sprintf("# SiteComponent: %s", n.Identifier()),
	}

	// Render the terraform config
	val, err := renderSiteComponentTerraformConfig(cfg, n)
	if err != nil {
		return "", fmt.Errorf("renderSiteTerraformConfig: %w", err)
	}
	result = append(result, val)

	// Render all the file sources
	val, err = renderFileSources(cfg, n.SiteConfig)
	if err != nil {
		return "", fmt.Errorf("failed to render file sources: %w", err)
	}
	result = append(result, val)

	// Render all the resources required by the site siteComponent
	val, err = renderSiteComponentResources(cfg, n)
	if err != nil {
		return "", fmt.Errorf("failed to render resources: %w", err)
	}
	result = append(result, val)

	// Render data links to other deployments
	val, err = renderRemoteSources(cfg, n)
	if err != nil {
		return "", fmt.Errorf("failed to render remote sources: %w", err)
	}
	result = append(result, val)

	// Render the siteComponent module
	val, err = renderComponentModule(ctx, cfg, n)
	if err != nil {
		return "", fmt.Errorf("failed to render component: %w", err)
	}
	result = append(result, val)

	return strings.Join(result, "\n"), nil
}

// renderSiteComponentTerraformConfig uses templates/terraform.tmpl to generate a terraform snippet for each component
func renderSiteComponentTerraformConfig(cfg *config.MachConfig, n graph.Node) (string, error) {
	tpl, err := templates.ReadFile("templates/terraform.tmpl")
	if err != nil {
		return "", err
	}

	site := n.(*graph.SiteComponent).SiteConfig
	siteComponent := n.(*graph.SiteComponent).SiteComponentConfig

	var providers []string
	for _, plugin := range cfg.Plugins.Names(siteComponent.Definition.Integrations...) {
		content, err := plugin.RenderTerraformProviders(site.Identifier)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render providers: %w", plugin.Name, err)
		}
		if content != "" {
			providers = append(providers, content)
		}
	}

	s, ok := cfg.StateRepository.Get(n.Identifier())
	if !ok {
		return "", fmt.Errorf("state repository does not have a backend for site %s", site.Identifier)
	}

	bc, err := s.Backend()
	if err != nil {
		return "", err
	}

	templateContext := struct {
		Providers     []string
		BackendConfig string
		IncludeSOPS   bool
	}{
		Providers:     providers,
		BackendConfig: bc,
		IncludeSOPS:   cfg.Variables.HasEncrypted(site.Identifier),
	}
	return utils.RenderGoTemplate(string(tpl), templateContext)
}

// renderSiteComponentResources uses templates/resources.tmpl to generate a terraform snippet for each component
func renderSiteComponentResources(cfg *config.MachConfig, n *graph.SiteComponent) (string, error) {
	tpl, err := templates.ReadFile("templates/resources.tmpl")
	if err != nil {
		return "", err
	}

	var resources []string
	for _, plugin := range cfg.Plugins.Names(n.SiteComponentConfig.Definition.Integrations...) {
		content, err := plugin.RenderTerraformResources(n.SiteConfig.Identifier)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render resources: %w", plugin.Name, err)
		}

		if content != "" {
			resources = append(resources, content)
		}
	}

	return utils.RenderGoTemplate(string(tpl), resources)
}

// renderComponentModule uses templates/component.tmpl to generate a terraform snippet for each component
func renderComponentModule(_ context.Context, cfg *config.MachConfig, n *graph.SiteComponent) (string, error) {
	tpl, err := templates.ReadFile("templates/site_component.tmpl")
	if err != nil {
		return "", err
	}

	tc := componentContext{
		ComponentName:    n.SiteComponentConfig.Name,
		ComponentVersion: n.SiteComponentConfig.Definition.Version,
		SiteName:         n.SiteConfig.Identifier,
		Environment:      cfg.Global.Environment,
		PluginResources:  []string{},
		PluginVariables:  []string{},
		PluginDependsOn:  []string{},
		PluginProviders:  []string{},
	}

	for _, plugin := range cfg.Plugins.Names(n.SiteComponentConfig.Definition.Integrations...) {
		plugin, err := cfg.Plugins.Get(plugin.Name)
		if err != nil {
			return "", err
		}

		cr, err := plugin.RenderTerraformComponent(n.SiteConfig.Identifier, n.SiteComponentConfig.Name)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render siteComponent: %w", plugin.Name, err)
		}

		if cr == nil {
			continue
		}

		tc.PluginResources = append(tc.PluginResources, cr.Resources)
		tc.PluginVariables = append(tc.PluginVariables, cr.Variables)
		tc.PluginProviders = append(tc.PluginProviders, cr.Providers...)
		tc.PluginDependsOn = append(tc.PluginDependsOn, cr.DependsOn...)
	}

	if n.SiteComponentConfig.HasCloudIntegration(&cfg.Global) {
		tc.HasCloudIntegration = true
		tc.ComponentVariables = "variables = {}"
		tc.ComponentSecrets = "secrets = {}"
	}

	if len(n.SiteComponentConfig.Variables) > 0 {
		val, err := serializeToHCL("variables", n.SiteComponentConfig.Variables, n.SiteComponentConfig.Deployment.Type, cfg.StateRepository, n.SiteConfig.Identifier)
		if err != nil {
			return "", err
		}
		tc.ComponentVariables = val
	}
	if len(n.SiteComponentConfig.Secrets) > 0 {
		val, err := serializeToHCL("secrets", n.SiteComponentConfig.Secrets, n.SiteComponentConfig.Deployment.Type, cfg.StateRepository, n.SiteConfig.Identifier)
		if err != nil {
			return "", err
		}
		tc.ComponentSecrets = val
	}

	vs, err := n.SiteComponentConfig.Definition.Source.GetVersionSource(n.SiteComponentConfig.Definition.Version)
	if err != nil {
		return "", err
	}

    // Escape backslashes in paths (Windows path separator)
    if runtime.GOOS == "windows" {
        tc.Source = strings.Replace(vs, "\\", "\\\\", -1)
    } else {
        tc.Source = vs
    }

	val, err := utils.RenderGoTemplate(string(tpl), tc)
	if err != nil {
		return "", fmt.Errorf("failed rendering site component: %w", err)
	}
	return val, nil
}

// renderRemoteSources uses the state repository to generate a terraform remote_state snippet for each referenced component
func renderRemoteSources(cfg *config.MachConfig, n *graph.SiteComponent) (string, error) {
	siteComponentConfig := n.SiteComponentConfig
	siteConfig := n.SiteConfig

	parents := append(
		siteComponentConfig.Variables.ListReferencedComponents(),
		siteComponentConfig.Secrets.ListReferencedComponents()...,
	)

	var links []string
	for _, parent := range parents {
		key, ok := cfg.StateRepository.StateKey(graph.CreateIdentifier(siteConfig.Identifier, parent))
		if !ok {
			return "", fmt.Errorf("missing remote state for %s", parent)
		}
		links = append(links, key)
	}

	links = slices.Compact(links)

	var result []string

	for _, link := range links {
		var linkIdentifier string
		if siteConfig.Identifier == link {
			linkIdentifier = link
		} else {
			linkIdentifier = graph.CreateIdentifier(siteConfig.Identifier, link)
		}

		s, ok := cfg.StateRepository.Get(linkIdentifier)
		if !ok {
			return "", fmt.Errorf("state repository does not have a backend for %s", link)
		}

		rs, err := s.RemoteState()
		if err != nil {
			return "", err
		}

		result = append(result, rs)
	}

	return strings.Join(result, "\n"), nil
}
