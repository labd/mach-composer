package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/elliotchance/pie/v2"
	"gopkg.in/yaml.v3"

	"github.com/labd/mach-composer/internal/utils"
)

// parseComponentsNode parses the `components:` block in the config file. It
// can also load the components from an external file using the ${include()}
// syntax.
func parseComponentsNode(cfg *MachConfig, node *yaml.Node, source string) error {
	if node.Tag == "!!str" {
		path := filepath.Dir(source)
		var err error
		node, err = loadComponentsNode(node, path)
		if err != nil {
			return err
		}
	}

	if err := node.Decode(&cfg.Components); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	if err := verifyComponents(cfg); err != nil {
		return fmt.Errorf("verify of components failed: %w", err)
	}

	if err := registerComponentEndpoints(cfg); err != nil {
		return fmt.Errorf("verify of components failed: %w", err)
	}

	knownKeys := []string{
		"name", "source", "version", "branch", "integrations", "endpoints",
	}
	for _, component := range node.Content {
		nodes := mapYamlNodes(component.Content)
		identifier := nodes["name"].Value
		err := iterateYamlNodes(nodes, knownKeys, func(key string, data map[string]any) error {
			return cfg.Plugins.SetComponentConfig(key, identifier, data)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func registerComponentEndpoints(cfg *MachConfig) error {
	cloudPlugin, err := cfg.Plugins.Get(cfg.Global.Cloud)
	if err != nil {
		return err
	}

	for i := range cfg.Components {
		c := &cfg.Components[i]
		err := cloudPlugin.SetComponentEndpointsConfig(c.Name, c.Endpoints)
		if err != nil {
			return err
		}
	}
	return err
}

// Verify the components config and set default values where needed.
func verifyComponents(cfg *MachConfig) error {
	seen := []string{}
	for i := range cfg.Components {
		c := &cfg.Components[i]

		// Make sure the component names are unique. Otherwise raise an error
		if pie.Contains(seen, c.Name) {
			return fmt.Errorf("component %s is duplicate", c.Name)
		}

		// If the component has no integrations (or now called plugins)
		// specified then set it to the cloud integration
		if len(c.Integrations) < 1 {
			c.Integrations = append(c.Integrations, cfg.Global.Cloud)
		}

		// If the source is a relative locale path then transform it to an
		// absolute path (required for Terraform)
		if strings.HasPrefix(c.Source, ".") {
			if val, err := filepath.Abs(c.Source); err == nil {
				c.Source = val
			} else {
				return err
			}
		}

		seen = append(seen, c.Name)
	}

	return nil
}

func loadComponentsNode(node *yaml.Node, path string) (*yaml.Node, error) {
	re := regexp.MustCompile(`\$\{include\(([^)]+)\)\}`)
	data := re.FindStringSubmatch(node.Value)
	if len(data) != 2 {
		return nil, fmt.Errorf("failed to parse ${include()} tag")
	}
	filename := filepath.Join(path, data[1])
	body, err := utils.AFS.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	result := yaml.Node{}
	if err = yaml.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Content) != 1 {
		return nil, fmt.Errorf("Invalid yaml file")
	}
	return result.Content[0], nil
}
