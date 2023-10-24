package generator

import (
	"context"
	"embed"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/utils"
	"strings"
)

//go:embed templates/component.tmpl
var componentTpl embed.FS

type componentContext struct {
	ComponentName       string
	SiteName            string
	Environment         string
	Source              string
	PluginResources     []string
	PluginProviders     []string
	PluginDependsOn     []string
	PluginVariables     []string
	ComponentVariables  string
	ComponentSecrets    string
	ComponentVersion    string
	HasCloudIntegration bool
}

// renderComponent uses templates/component.tf to generate a terraform snippet for each component
func renderComponent(_ context.Context, cfg *config.MachConfig, site *config.SiteConfig, component *config.SiteComponentConfig) (string, error) {
	result := []string{
		"# This file is auto-generated by MACH composer",
		fmt.Sprintf("# Component: %s", component.Name),
	}

	// Render the terraform config
	val, err := renderTerraformConfig(cfg, site, component.Name)
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
	val, err = renderResources(cfg, site)
	if err != nil {
		return "", fmt.Errorf("failed to render resources: %w", err)
	}
	result = append(result, val)

	// Render all the global resources
	val, err = renderRemoteSources(cfg, component)
	if err != nil {
		return "", fmt.Errorf("failed to render remote sources: %w", err)
	}
	result = append(result, val)

	tc := componentContext{
		ComponentName:    component.Name,
		ComponentVersion: component.Definition.Version,
		SiteName:         site.Identifier,
		Environment:      cfg.Global.Environment,
		Source:           component.Definition.Source,
		PluginResources:  []string{},
		PluginVariables:  []string{},
		PluginDependsOn:  []string{},
		PluginProviders:  []string{},
	}

	for _, plugin := range cfg.Plugins.All() {
		if !pie.Contains(component.Definition.Integrations, plugin.Name) {
			continue
		}
		plugin, err := cfg.Plugins.Get(plugin.Name)
		if err != nil {
			return "", err
		}

		cr, err := plugin.RenderTerraformComponent(site.Identifier, component.Name)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render component: %w", plugin.Name, err)
		}

		if cr == nil {
			continue
		}

		tc.PluginResources = append(tc.PluginResources, cr.Resources)
		tc.PluginVariables = append(tc.PluginVariables, cr.Variables)
		tc.PluginProviders = append(tc.PluginProviders, cr.Providers...)
		tc.PluginDependsOn = append(tc.PluginDependsOn, cr.DependsOn...)
	}

	tpl, err := componentTpl.ReadFile("templates/component.tmpl")
	if err != nil {
		return "", err
	}

	if component.HasCloudIntegration(&cfg.Global) {
		tc.HasCloudIntegration = true
		tc.ComponentVariables = "variables = {}"
		tc.ComponentSecrets = "secrets = {}"
	}

	if len(component.Variables) > 0 {
		val, err := serializeToHCL("variables", component.Variables)
		if err != nil {
			return "", err
		}
		tc.ComponentVariables = val
	}
	if len(component.Secrets) > 0 {
		val, err := serializeToHCL("secrets", component.Secrets)
		if err != nil {
			return "", err
		}
		tc.ComponentSecrets = val
	}

	if component.Definition.IsGitSource() {
		// When using Git, we will automatically add a reference to the string
		// so that the given version is used when fetching the module itself
		// from Git as well
		tc.Source += fmt.Sprintf("?ref=%s", component.Definition.Version)
	}

	val, err = utils.RenderGoTemplate(string(tpl), tc)
	if err != nil {
		return "", fmt.Errorf("renderTerraformConfig: %w", err)
	}
	result = append(result, val)

	content := strings.Join(result, "\n")
	return content, nil
}

func renderRemoteSources(cfg *config.MachConfig, component *config.SiteComponentConfig) (string, error) {
	parents := component.Variables.ListComponents()
	parents = append(parents, component.Secrets.ListComponents()...)

	var result []string

	for _, parent := range parents {
		if !cfg.StateRepository.Has(parent) {
			return "", fmt.Errorf("missing remoteState for %s", parent)
		}

		remoteState, err := cfg.StateRepository.Get(parent).RemoteState()
		if err != nil {
			return "", err
		}

		result = append(result, remoteState)
	}

	return strings.Join(result, "\n"), nil
}
