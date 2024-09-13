package graph

import (
	"github.com/dominikbraun/graph"
)

type Path []string

func fetchPathsToTarget(source, target string, pm map[string]map[string]graph.Edge[string], currentPath Path) []Path {
	var paths []Path
	parents := pm[source]
	if len(parents) == 0 {
		return []Path{currentPath}
	}

	currentPath = append(currentPath, source)

	for _, parent := range parents {
		if parent.Source == target {
			paths = []Path{currentPath}
		}
		newPaths := fetchPathsToTarget(parent.Source, target, pm, currentPath)
		paths = append(paths, newPaths...)
	}

	return paths
}
