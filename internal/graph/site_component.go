package graph

import (
	"github.com/dominikbraun/graph"
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"sort"
)

type SiteComponent struct {
	baseNode
	ProjectConfig       config.MachConfig
	SiteConfig          config.SiteConfig
	SiteComponentConfig config.SiteComponentConfig
}

func NewSiteComponent(
	g graph.Graph[string, Node], path, identifier string, deploymentType config.DeploymentType,
	ancestor Node, projectConfig config.MachConfig, siteConfig config.SiteConfig, siteComponentConfig config.SiteComponentConfig,
) *SiteComponent {
	return &SiteComponent{
		baseNode:            newBaseNode(g, path, identifier, SiteComponentType, ancestor, deploymentType, true),
		ProjectConfig:       projectConfig,
		SiteConfig:          siteConfig,
		SiteComponentConfig: siteComponentConfig,
	}
}

func (sc *SiteComponent) Hash() (string, error) {
	return HashSiteComponent(sc)
}

func SortSiteComponentNodes(nodes []*SiteComponent) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].SiteComponentConfig.Name < nodes[j].SiteComponentConfig.Name
	})
}
