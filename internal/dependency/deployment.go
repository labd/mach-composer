package dependency

import (
	"errors"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/mach-composer/mach-composer-cli/internal/config"
)

func ToDeploymentGraph(cfg *config.MachConfig) (*Graph, error) {
	g, err := ToDependencyGraph(cfg)
	if err != nil {
		return nil, err
	}

	if err := validateDeployment(g); err != nil {
		return nil, err
	}

	// Reduce all nodes that are not independent to site node
	if err = reduceNodes(g); err != nil {
		return nil, err
	}

	return g, nil
}

func reduceNodes(g *Graph) error {
	var pErr error
	if err := graph.BFS(g.NodeGraph, g.StartNode.Path(), func(p string) bool {
		n, _ := g.Vertex(p)

		if !n.Independent() {
			siteNode, ok := n.Parent().(*Site)
			if !ok {
				pErr = fmt.Errorf("node %s is expected to have site as parent", n.Path())
				return true
			}

			siteComponentNode, ok := n.(*SiteComponent)
			if !ok {
				pErr = fmt.Errorf("node %s is expected to be a site component", n.Path())
				return true
			}

			var siteComponentConfig = siteComponentNode.SiteComponentConfig

			siteNode.NestedSiteComponentConfigs = append(siteNode.NestedSiteComponentConfigs, siteComponentConfig)

			am, _ := g.AdjacencyMap()
			pm, _ := g.PredecessorMap()

			childEdges := am[p]
			parentEdges := pm[p]

			for _, childEdge := range childEdges {
				if err := g.RemoveEdge(childEdge.Source, childEdge.Target); err != nil {
					pErr = err
					return false
				}
			}

			for _, parentEdge := range parentEdges {
				if err := g.RemoveEdge(parentEdge.Source, parentEdge.Target); err != nil {
					pErr = err
					return false
				}
			}

			for _, childEdge := range childEdges {
				for _, parentEdge := range parentEdges {

					_, err := g.Edge(parentEdge.Source, childEdge.Target)
					if err != nil && !errors.Is(err, graph.ErrEdgeNotFound) {
						pErr = err
						return false
					}

					if err != nil && errors.Is(err, graph.ErrEdgeNotFound) {
						if err := g.AddEdge(parentEdge.Source, childEdge.Target); err != nil {
							pErr = err
							return false
						}
					}
				}
			}

			if err := g.RemoveVertex(n.Path()); err != nil {
				pErr = err
				return true
			}
		}

		return false
	}); err != nil {
		return err
	}

	return pErr
}