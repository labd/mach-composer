package graph

import (
	"github.com/dominikbraun/graph"
	"golang.org/x/exp/maps"
)

type Graph struct {
	graph.Graph[string, Node]
	StartNode Node
}

type Vertices []Node

// Vertices returns all the vertex that are contained in the graph
func (g *Graph) Vertices() Vertices {
	var vertices Vertices

	m, _ := g.AdjacencyMap()

	keys := maps.Keys(m)

	for _, k := range keys {
		v, _ := g.Vertex(k)
		vertices = append(vertices, v)
	}

	return vertices
}

func (g *Graph) VerticesByType(typ Type) Vertices {
	v := g.Vertices()
	var vertices Vertices

	for _, node := range v {
		if node.Type() == typ {
			vertices = append(vertices, node)
		}
	}

	return vertices
}

func (g *Graph) VertexByIdentifier(identifier string) Node {
	for _, v := range g.Vertices() {
		if v.Identifier() == identifier {
			return v
		}
	}
	return nil
}

func (g *Graph) VertexExists(identifier string) bool {
	return g.VertexByIdentifier(identifier) != nil
}

// Routes determines all the possible paths between two nodes
func (g *Graph) Routes(source, target string) ([]Path, error) {
	var routes []Path

	m, err := g.PredecessorMap()
	if err != nil {
		return routes, err
	}

	eg := m[source]

	for _, pathElement := range eg {
		p := []string{pathElement.Source}
		newRoutes := fetchPathsToTarget(pathElement.Source, target, m, p)
		routes = append(routes, newRoutes...)
	}

	return routes, nil
}
