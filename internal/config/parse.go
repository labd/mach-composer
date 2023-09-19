package config

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mach-composer/mach-composer-cli/internal/state"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/mach-composer/mach-composer-cli/internal/plugins"
	"github.com/mach-composer/mach-composer-cli/internal/utils"
	"github.com/mach-composer/mach-composer-cli/internal/variables"
)

type ConfigOptions struct {
	VarFilenames []string
	Plugins      *plugins.PluginRepository

	NoResolveVars bool
}

// Open is the main entrypoint for this module. It opens the given yaml filename
// and reads it to construct the MachConfig.
// Note that you need to close the MachConfig via the Close() method in order
// to clean up.
func Open(ctx context.Context, filename string, opts *ConfigOptions) (*MachConfig, error) {
	var pluginRepo *plugins.PluginRepository
	if opts != nil {
		pluginRepo = opts.Plugins
	}

	raw, err := loadConfig(ctx, filename, pluginRepo)
	if err != nil {
		return nil, err
	}

	// Validate again
	isValid, err := validateCompleteConfig(raw)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("failed to load config %s due to errors", filename)
	}

	for _, f := range opts.VarFilenames {
		if err := raw.variables.Load(ctx, f); err != nil {
			return nil, err
		}
	}

	// For some actions we don't want to resolve variables since they then need
	// to be passed as argument.
	if !opts.NoResolveVars {
		if err := resolveVariables(ctx, raw); err != nil {
			if notFoundErr, ok := err.(*variables.NotFoundError); ok {
				err = &SyntaxError{
					message:  fmt.Sprintf("unable to resolve variable %#v", notFoundErr.Name),
					line:     notFoundErr.Node.Line,
					filename: raw.filename,
					column:   notFoundErr.Node.Column,
				}
			}
			return nil, err
		}
	}

	return resolveConfig(ctx, raw)
}

func loadConfig(ctx context.Context, filename string, pr *plugins.PluginRepository) (*rawConfig, error) {
	// Load the yaml file and do basic validation if the config file is valid
	// based on a json schema
	document, err := loadYamlFile(filename)
	if err != nil {
		return nil, err
	}

	// Initial validation. We validate the document twice, once only the
	// structure and later again when we loaded the plugins
	isValid, err := validateConfig(document)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("failed to load config %s due to errors", filename)
	}

	// Decode the yaml in an intermediate config file
	raw, err := newRawConfig(filename, document)
	if err != nil {
		return nil, err
	}

	if _, err := LoadRefData(ctx, &raw.Components); err != nil {
		return nil, err
	}

	// Load the plugins
	raw.plugins = pr
	if err := loadPlugins(ctx, raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func loadPlugins(ctx context.Context, raw *rawConfig) error {
	if raw.plugins != nil {
		return nil
	}
	raw.plugins = plugins.NewPluginRepository()

	if len(raw.MachComposer.Plugins) == 0 {
		log.Debug().Msg("No plugins specified; loading default plugins")
		raw.MachComposer.Plugins = map[string]MachPluginConfig{
			"amplience": {
				Source:  "mach-composer/amplience",
				Version: "0.1.3",
			},
			"aws": {
				Source:  "mach-composer/aws",
				Version: "0.1.0",
			},
			"azure": {
				Source:  "mach-composer/azure",
				Version: "0.1.0",
			},
			"commercetools": {
				Source:  "mach-composer/commercetools",
				Version: "0.1.5",
			},
			"contentful": {
				Source:  "mach-composer/contentful",
				Version: "0.1.0",
			},
			"sentry": {
				Source:  "mach-composer/sentry",
				Version: "0.1.2",
			},
		}
	}

	for pluginName, pluginData := range raw.MachComposer.Plugins {
		pluginConfig := plugins.PluginConfig{
			Source:  pluginData.Source,
			Version: pluginData.Version,
		}
		if err := raw.plugins.LoadPlugin(ctx, pluginName, pluginConfig); err != nil {
			return err
		}
	}
	return nil
}

// parseConfig is responsible for parsing a mach composer yaml config file and
// creating the resulting MachConfig struct.
func resolveConfig(ctx context.Context, intermediate *rawConfig) (*MachConfig, error) {
	if err := intermediate.validate(); err != nil {
		return nil, err
	}

	cfgHash, err := intermediate.computeHash()
	if err != nil {
		return nil, err
	}

	cfg := &MachConfig{
		ConfigHash:      cfgHash,
		StateRepository: state.NewRepository(),
		extraFiles:      make(map[string][]byte),
		Filename:        filepath.Base(intermediate.filename),
		MachComposer:    intermediate.MachComposer,
		Variables:       intermediate.variables,
		Plugins:         intermediate.plugins,
	}

	if err := parseGlobalNode(cfg, &intermediate.Global); err != nil {
		if _, ok := err.(*plugins.PluginNotFoundError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to parse global node: %w", err)
	}

	if err := parseComponentsNode(cfg, &intermediate.Components, intermediate.filename); err != nil {
		return nil, fmt.Errorf("failed to parse components node: %w", err)
	}

	if err := parseSitesNode(cfg, &intermediate.Sites); err != nil {
		return nil, fmt.Errorf("failed to parse sites node: %w", err)
	}

	return cfg, nil
}

func resolveVariables(ctx context.Context, rawConfig *rawConfig) error {
	vars := rawConfig.variables

	if rawConfig.MachComposer.VariablesFile != "" {
		if err := vars.Load(ctx, rawConfig.MachComposer.VariablesFile); err != nil {
			return err
		}
	}

	if err := vars.InterpolateNode(&rawConfig.Global); err != nil {
		return err
	}

	if err := vars.InterpolateNode(&rawConfig.Components); err != nil {
		return err
	}

	// Interpolate the variables per-site to keep track of which site uses which
	// variable.
	for _, node := range rawConfig.Sites.Content {
		mapping := mapYamlNodes(node.Content)
		if idNode, ok := mapping["identifier"]; ok {
			siteId := idNode.Value
			if err := vars.InterpolateSiteNode(siteId, node); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadYamlFile(filename string) (*yaml.Node, error) {
	// Read the config file from the given filename
	body, err := utils.AFS.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Read the yaml nodes
	document := &yaml.Node{}
	if err := yaml.Unmarshal(body, document); err != nil {
		return nil, err
	}
	return document, nil
}
