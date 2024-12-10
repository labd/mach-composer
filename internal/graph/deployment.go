package graph

import (
	"errors"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/mach-composer/mach-composer-cli/internal/config"
)

type options struct {
	siteTarget string
}

type Option func(o *options)

func WithTargetSiteName(site string) Option {
	return func(o *options) {
		o.siteTarget = site
	}
}

// ToDeploymentGraph converts a MachConfig to a Graph ready for deployment.
// This means that all nodes that are not independently deployable are pruned from the graph.
func ToDeploymentGraph(cfg *config.MachConfig, outPath string, opts ...Option) (*Graph, error) {
	o := options{
		siteTarget: "",
	}

	for _, opt := range opts {
		opt(&o)
	}

	g, err := ToDependencyGraph(cfg, outPath)
	if err != nil {
		return nil, err
	}

	if err := validateDeployment(g); err != nil {
		return nil, err
	}

	// Remove all nodes that are not independent to site node
	if err := reduceNodes(g); err != nil {
		return nil, err
	}

	//Prune to only include the site node if provided
	if err := targetSiteNode(g, o.siteTarget); err != nil {
		return nil, err
	}

	return g, nil
}

func validateDeployment(g *Graph) error {
	var errList errorList
	err := graph.DFS(g.Graph, g.StartNode.Path(), func(p string) bool {
		n, _ := g.Vertex(p)

		am, _ := g.AdjacencyMap()
		edges := am[p]

		for _, edge := range edges {
			child, _ := g.Vertex(edge.Target)
			if n.Type() == SiteComponentType && n.Independent() && !child.Independent() {
				errList.AddError(fmt.Errorf("baseNode %s is independent but has a dependent child %s", n.Path(), child.Path()))
			}
		}

		return false
	})
	if err != nil {
		return err
	}
	if len(errList) > 0 {
		return &ValidationError{
			Msg:    "validation failed",
			Errors: errList,
		}
	}

	return nil
}

func reduceNodes(g *Graph) error {
	var pErr error
	if err := graph.BFS(g.Graph, g.StartNode.Path(), func(p string) bool {
		n, _ := g.Graph.Vertex(p)

		if !n.Independent() {
			siteNode, ok := n.Ancestor().(*Site)
			if !ok {
				pErr = fmt.Errorf("node %s is expected to have site as parent", n.Path())
				return true
			}

			siteComponentNode, ok := n.(*SiteComponent)
			if !ok {
				pErr = fmt.Errorf("node %s is expected to be a site component", n.Path())
				return true
			}

			siteNode.NestedNodes = append(siteNode.NestedNodes, siteComponentNode)

			am, _ := g.Graph.AdjacencyMap()
			pm, _ := g.Graph.PredecessorMap()

			childEdges := am[p]
			parentEdges := pm[p]

			for _, childEdge := range childEdges {
				if err := g.Graph.RemoveEdge(childEdge.Source, childEdge.Target); err != nil {
					pErr = err
					return false
				}
			}

			for _, parentEdge := range parentEdges {
				if err := g.Graph.RemoveEdge(parentEdge.Source, parentEdge.Target); err != nil {
					pErr = err
					return false
				}
			}

			for _, childEdge := range childEdges {
				for _, parentEdge := range parentEdges {

					_, err := g.Graph.Edge(parentEdge.Source, childEdge.Target)
					if err != nil && !errors.Is(err, graph.ErrEdgeNotFound) {
						pErr = err
						return false
					}

					if err != nil && errors.Is(err, graph.ErrEdgeNotFound) {
						if err := g.Graph.AddEdge(parentEdge.Source, childEdge.Target); err != nil {
							pErr = err
							return false
						}
					}
				}
			}

			if err := g.Graph.RemoveVertex(n.Path()); err != nil {
				pErr = err
				return true
			}
		}

		return false
	}); err != nil {
		return err
	}

	for _, v := range g.Vertices() {
		v.resetGraph(g.Graph)
	}

	return pErr
}

func targetSiteNode(g *Graph, site string) error {
	if site == "" {
		return nil
	}

	if !g.VertexExists(site) {
		return fmt.Errorf("site node %s does not exist", site)
	}

	for _, v := range g.VerticesByType(SiteType) {
		if v.Identifier() != site {
			var pErr error
			_ = graph.DFS(g.Graph, v.Path(), func(p string) bool {
				n, err := g.Graph.Vertex(p)
				if err != nil {
					pErr = err
					return true
				}

				n.SetTargeted(false)

				return false
			})

			return pErr
		}
	}

	return nil
}
