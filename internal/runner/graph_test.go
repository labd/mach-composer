package runner

import (
	"context"
	"errors"
	"github.com/mach-composer/mach-composer-cli/internal/batcher"
	"github.com/mach-composer/mach-composer-cli/internal/cli"
	internalgraph "github.com/mach-composer/mach-composer-cli/internal/graph"
	"github.com/mach-composer/mach-composer-cli/internal/hash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGraphRunnerMultipleLevels(t *testing.T) {
	project := new(internalgraph.NodeMock)
	project.On("Identifier").Return("main")
	project.On("Path").Return("main")
	project.On("Hash").Return("main", nil)
	project.On("Type").Return(internalgraph.ProjectType)

	site := new(internalgraph.NodeMock)
	site.On("Identifier").Return("site-1")
	site.On("Path").Return("site-1")
	site.On("Hash").Return("site-1", nil)
	site.On("Type").Return(internalgraph.SiteType)
	site.On("Targeted").Return(true)

	component1 := new(internalgraph.NodeMock)
	component1.On("Identifier").Return("component-1")
	component1.On("Path").Return("component-1")
	component1.On("Hash").Return("component-1", nil)
	component1.On("Type").Return(internalgraph.SiteComponentType)
	component1.On("Targeted").Return(true)

	component2 := new(internalgraph.NodeMock)
	component2.On("Identifier").Return("component-2")
	component2.On("Path").Return("component-2")
	component2.On("Hash").Return("component-2", nil)
	component2.On("Type").Return(internalgraph.SiteComponentType)
	component2.On("Targeted").Return(true)

	component3 := new(internalgraph.NodeMock)
	component3.On("Identifier").Return("component-3")
	component3.On("Path").Return("component-3")
	component3.On("Hash").Return("component-3", nil)
	component3.On("Type").Return(internalgraph.SiteComponentType)
	component3.On("Targeted").Return(true)

	graph := internalgraph.CreateGraphMock(
		map[string]internalgraph.Node{
			"main":        project,
			"site-1":      site,
			"component-1": component1,
			"component-2": component2,
			"component-3": component3,
		},
		project,
		internalgraph.EdgeMock{Source: "main", Target: "site-1"},
		internalgraph.EdgeMock{Source: "site-1", Target: "component-1"},
		internalgraph.EdgeMock{Source: "site-1", Target: "component-2"},
		internalgraph.EdgeMock{Source: "component-1", Target: "component-3"},
		internalgraph.EdgeMock{Source: "component-2", Target: "component-3"},
	)

	runner := GraphRunner{workers: 1}
	runner.hash = hash.NewMemoryMapHandler(
		hash.Entry{Identifier: "site-1", Hash: "site-1"},
		hash.Entry{Identifier: "component-1", Hash: "component-1"},
	)
	runner.batch = batcher.NaiveBatchFunc()

	var called []string

	err := runner.run(context.Background(), graph, func(ctx context.Context, node internalgraph.Node) error {
		called = append(called, node.Identifier())
		return nil
	}, false)

	assert.NoError(t, err)
	assert.Equal(t, []string{"component-2", "component-3"}, called)
}

func TestGraphRunnerTargeted(t *testing.T) {
	project := new(internalgraph.NodeMock)
	project.On("Identifier").Return("main")
	project.On("Path").Return("main")
	project.On("Hash").Return("main", nil)
	project.On("Type").Return(internalgraph.ProjectType)

	site := new(internalgraph.NodeMock)
	site.On("Identifier").Return("site-1")
	site.On("Path").Return("site-1")
	site.On("Hash").Return("site-1", nil)
	site.On("Type").Return(internalgraph.SiteType)
	site.On("Targeted").Return(false)

	graph := internalgraph.CreateGraphMock(
		map[string]internalgraph.Node{
			"main":   project,
			"site-1": site,
		},
		project,
		internalgraph.EdgeMock{Source: "main", Target: "site-1"},
	)

	runner := GraphRunner{workers: 1}
	runner.hash = hash.NewMemoryMapHandler()
	runner.batch = batcher.NaiveBatchFunc()

	var called []string

	err := runner.run(context.Background(), graph, func(ctx context.Context, node internalgraph.Node) error {
		called = append(called, node.Identifier())
		return nil
	}, false)

	assert.NoError(t, err)
	assert.Len(t, called, 0)
}

func TestGraphRunnerError(t *testing.T) {
	project := new(internalgraph.NodeMock)
	project.On("Identifier").Return("main")
	project.On("Path").Return("main")
	project.On("Hash").Return("main", nil)
	project.On("Type").Return(internalgraph.ProjectType)

	site := new(internalgraph.NodeMock)
	site.On("Identifier").Return("site-1")
	site.On("Path").Return("site-1")
	site.On("Hash").Return("site-1", nil)
	site.On("Type").Return(internalgraph.SiteType)
	site.On("Targeted").Return(true)

	component1 := new(internalgraph.NodeMock)
	component1.On("Identifier").Return("component-1")
	component1.On("Path").Return("component-1")
	component1.On("Hash").Return("component-1", nil)
	component1.On("Type").Return(internalgraph.SiteComponentType)
	component1.On("Targeted").Return(true)

	component2 := new(internalgraph.NodeMock)
	component2.On("Identifier").Return("component-2")
	component2.On("Path").Return("component-2")
	component2.On("Hash").Return("component-2", nil)
	component2.On("Type").Return(internalgraph.SiteComponentType)
	component2.On("Targeted").Return(true)

	component3 := new(internalgraph.NodeMock)
	component3.On("Identifier").Return("component-3")
	component3.On("Path").Return("component-3")
	component3.On("Hash").Return("component-3", nil)
	component3.On("Type").Return(internalgraph.SiteComponentType)
	component3.On("Targeted").Return(true)

	graph := internalgraph.CreateGraphMock(
		map[string]internalgraph.Node{
			"main":        project,
			"site-1":      site,
			"component-1": component1,
			"component-2": component2,
			"component-3": component3,
		},
		project,
		internalgraph.EdgeMock{Source: "main", Target: "site-1"},
		internalgraph.EdgeMock{Source: "site-1", Target: "component-1"},
		internalgraph.EdgeMock{Source: "site-1", Target: "component-2"},
		internalgraph.EdgeMock{Source: "component-1", Target: "component-3"},
		internalgraph.EdgeMock{Source: "component-2", Target: "component-3"},
	)

	runner := GraphRunner{workers: 1}
	runner.hash = hash.NewMemoryMapHandler()
	runner.batch = batcher.NaiveBatchFunc()

	err := runner.run(context.Background(), graph, func(ctx context.Context, node internalgraph.Node) error {
		if node.Identifier() == "component-2" {
			return assert.AnError
		}
		return nil
	}, false)

	cliErr := &cli.GroupedError{}

	assert.ErrorAs(t, err, &cliErr)

	errors.As(err, &cliErr)
	assert.Len(t, cliErr.Errors, 1)
	assert.Equal(t, assert.AnError, cliErr.Errors[0])
}
