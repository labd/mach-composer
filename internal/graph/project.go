package graph

import (
	"github.com/dominikbraun/graph"
	"github.com/mach-composer/mach-composer-cli/internal/config"
)

type Project struct {
	baseNode
	ProjectConfig config.MachConfig
}

func NewProject(g graph.Graph[string, Node], path, identifier string, deploymentType config.DeploymentType,
	projectConfig config.MachConfig) *Project {
	return &Project{
		baseNode:      newBaseNode(g, path, identifier, ProjectType, nil, deploymentType, false),
		ProjectConfig: projectConfig,
	}
}

func (p *Project) Hash() (string, error) {
	return "", nil
}
