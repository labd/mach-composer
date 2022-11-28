package config

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/labd/mach-composer/internal/variables"
)

func parseSitesNode(cfg *MachConfig, sitesNode *yaml.Node) error {
	if err := sitesNode.Decode(&cfg.Sites); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	cloudPlugin, err := cfg.Plugins.Get(cfg.Global.Cloud)
	if err != nil {
		return err
	}

	knownKeys := []string{
		"name", "identifier", "endpoints", "components",
	}
	for _, site := range sitesNode.Content {
		nodes := mapYamlNodes(site.Content)
		siteId := nodes["identifier"].Value

		err := iterateYamlNodes(nodes, knownKeys, func(key string, data map[string]any) error {
			return cfg.Plugins.SetSiteConfig(key, siteId, data)
		})
		if err != nil {
			return err
		}

		if node, ok := nodes["endpoints"]; ok {
			data, err := nodeAsMap(node)
			if err != nil {
				return err
			}
			if err := cloudPlugin.SetSiteEndpointsConfig(siteId, data); err != nil {
				return err
			}
		}

		if err := parseSiteComponentsNode(cfg, siteId, nodes["components"]); err != nil {
			return err
		}
	}

	return resolveSiteComponents(cfg)
}

func parseSiteComponentsNode(cfg *MachConfig, site string, node *yaml.Node) error {
	knownKeys := []string{
		"name", "variables", "secrets",
		"store_variables", "store_secrets",
	}

	for _, component := range node.Content {
		nodes := mapYamlNodes(component.Content)
		identifier := nodes["name"].Value

		migrateCommercetools(site, identifier, nodes)

		for key, node := range nodes {
			if pie.Contains(knownKeys, key) {
				continue
			}
			data, err := nodeAsMap(node)
			if err != nil {
				return err
			}
			if err := cfg.Plugins.SetSiteComponentConfig(site, identifier, key, data); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveSiteComponents(cfg *MachConfig) error {
	components := make(map[string]*Component, len(cfg.Components))
	for i, c := range cfg.Components {
		components[c.Name] = &cfg.Components[i]
	}

	for _, site := range cfg.Sites {
		if len(site.Components) < 1 {
			continue
		}

		for i := range site.Components {
			c := &site.Components[i]

			ref, ok := components[c.Name]
			if !ok {
				return fmt.Errorf("component %s does not exist in global components.", c.Name)
			}
			c.Definition = ref

			var err error
			c.Variables, err = variables.InterpolateComponentVars(c.Variables)
			if err != nil {
				return err
			}

			c.Secrets, err = variables.InterpolateComponentVars(c.Secrets)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// migrateCommercetools moves the store_variables and store_secrets under the
// commercetools node. Needed to say backwards compatible
func migrateCommercetools(site, name string, nodes map[string]*yaml.Node) {
	needsMigrate := false
	if _, ok := nodes["store_variables"]; ok {
		needsMigrate = true
	}
	if _, ok := nodes["store_secrets"]; ok {
		needsMigrate = true
	}
	if !needsMigrate {
		return
	}

	logrus.Warnf("%s: %s store_variables and store_secrets should be children of the commercetools node\n", site, name)
	if _, ok := nodes["commercetools"]; !ok {
		nodes["commercetools"] = &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: []*yaml.Node{},
		}

		if val, ok := nodes["store_variables"]; ok {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "store_variables",
			}
			nodes["commercetools"].Content = append(
				nodes["commercetools"].Content,
				keyNode,
				val,
			)
		}
		if val, ok := nodes["store_secrets"]; ok {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "store_secrets",
			}
			nodes["commercetools"].Content = append(
				nodes["commercetools"].Content,
				keyNode,
				val,
			)
		}
	}
}
