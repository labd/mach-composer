package graph

import (
	"github.com/dominikbraun/graph"
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToDeploymentGraphSimple(t *testing.T) {
	cfg := &config.MachConfig{
		Filename: "main",
		MachComposer: config.MachComposer{
			Deployment: config.Deployment{
				Type: config.DeploymentSite,
			},
		},

		Sites: []config.SiteConfig{
			{
				Name: "site 1",
				Deployment: &config.Deployment{
					Type: config.DeploymentSite,
				},
				Identifier: "site-1",
				Components: []config.SiteComponentConfig{
					{
						Name: "site-component-1",
						Deployment: &config.Deployment{
							Type: config.DeploymentSite,
						},
					},
					{
						Name: "site-component-2",
						Deployment: &config.Deployment{
							Type: config.DeploymentSiteComponent,
						},
					},
				},
			},
		},
	}

	g, err := ToDeploymentGraph(cfg, "")
	assert.NoError(t, err)

	o, _ := g.Order()
	assert.Equal(t, 3, o)

	e, _ := g.Edges()
	assert.Len(t, e, 2)

	siteNode, err := g.Vertex("main/site-1")
	assert.NoError(t, err)
	assert.Equal(t, "main/site-1", siteNode.Path())
	assert.IsType(t, &Site{}, siteNode)
	assert.Equal(t, cfg.Sites[0], siteNode.(*Site).SiteConfig)
	assert.Equal(t, 1, len(siteNode.(*Site).NestedNodes))
	assert.Equal(t, "site-component-1", siteNode.(*Site).NestedNodes[0].SiteComponentConfig.Name)

	siteComponentNode, err := g.Vertex("main/site-1/site-component-2")
	assert.NoError(t, err)
	assert.IsType(t, &SiteComponent{}, siteComponentNode)

	am, _ := g.AdjacencyMap()
	assert.Equal(t, map[string]map[string]graph.Edge[string]{
		"main": {
			"main/site-1": graph.Edge[string]{
				Source: "main",
				Target: "main/site-1",
				Properties: graph.EdgeProperties{
					Attributes: map[string]string{},
					Weight:     0,
					Data:       nil,
				},
			},
		},
		"main/site-1": {
			"main/site-1/site-component-2": graph.Edge[string]{
				Source: "main/site-1",
				Target: "main/site-1/site-component-2",
				Properties: graph.EdgeProperties{
					Attributes: map[string]string{},
					Weight:     0,
					Data:       nil,
				},
			},
		},
		"main/site-1/site-component-2": {},
	}, am)
}

func TestTargetSiteNodeDoesNotExist(t *testing.T) {
	cfg := &config.MachConfig{
		Filename: "main",
		MachComposer: config.MachComposer{
			Deployment: config.Deployment{
				Type: config.DeploymentSite,
			},
		},

		Sites: []config.SiteConfig{
			{
				Name: "site 1",
				Deployment: &config.Deployment{
					Type: config.DeploymentSite,
				},
				Identifier: "site-1",
				Components: []config.SiteComponentConfig{
					{
						Name: "site-component-1",
						Deployment: &config.Deployment{
							Type: config.DeploymentSite,
						},
					},
				},
			},
		},
	}

	g, err := ToDeploymentGraph(cfg, "")
	assert.NoError(t, err)

	err = targetSiteNode(g, "site-2")
	assert.ErrorContains(t, err, "site-2 does not exist")
}

func TestTargetSiteNodeAll(t *testing.T) {
	cfg := &config.MachConfig{
		Filename: "main",
		MachComposer: config.MachComposer{
			Deployment: config.Deployment{
				Type: config.DeploymentSite,
			},
		},

		Sites: []config.SiteConfig{
			{
				Name: "site 1",
				Deployment: &config.Deployment{
					Type: config.DeploymentSite,
				},
				Identifier: "site-1",
				Components: []config.SiteComponentConfig{
					{
						Name: "site-component-1",
						Deployment: &config.Deployment{
							Type: config.DeploymentSite,
						},
					},
				},
			},
			{
				Name: "site 2",
				Deployment: &config.Deployment{
					Type: config.DeploymentSite,
				},
				Identifier: "site-2",
				Components: []config.SiteComponentConfig{
					{
						Name: "site-component-1",
						Deployment: &config.Deployment{
							Type: config.DeploymentSite,
						},
					},
				},
			},
		},
	}

	g, err := ToDeploymentGraph(cfg, "")
	assert.NoError(t, err)

	err = targetSiteNode(g, "")
	assert.NoError(t, err)

	var siteVertices = g.VerticesByType(SiteType)
	assert.Len(t, siteVertices, 2)

	for _, v := range siteVertices {
		assert.True(t, v.Targeted())
	}
}

func TestTargetSiteNodeTargeted(t *testing.T) {
	cfg := &config.MachConfig{
		Filename: "main",
		MachComposer: config.MachComposer{
			Deployment: config.Deployment{
				Type: config.DeploymentSite,
			},
		},

		Sites: []config.SiteConfig{
			{
				Name: "site 1",
				Deployment: &config.Deployment{
					Type: config.DeploymentSite,
				},
				Identifier: "site-1",
				Components: []config.SiteComponentConfig{
					{
						Name: "site-component-1",
						Deployment: &config.Deployment{
							Type: config.DeploymentSite,
						},
					},
				},
			},
			{
				Name: "site 2",
				Deployment: &config.Deployment{
					Type: config.DeploymentSite,
				},
				Identifier: "site-2",
				Components: []config.SiteComponentConfig{
					{
						Name: "site-component-1",
						Deployment: &config.Deployment{
							Type: config.DeploymentSite,
						},
					},
				},
			},
		},
	}

	g, err := ToDeploymentGraph(cfg, "")
	assert.NoError(t, err)

	err = targetSiteNode(g, "site-1")
	assert.NoError(t, err)

	var siteVertices = g.VerticesByType(SiteType)
	assert.Len(t, siteVertices, 2)

	site1 := g.VertexByIdentifier("site-1")
	assert.True(t, site1.Targeted())

	site2 := g.VertexByIdentifier("site-2")
	assert.False(t, site2.Targeted())
}
