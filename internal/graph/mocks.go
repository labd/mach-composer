package graph

import (
	"github.com/dominikbraun/graph"
	"github.com/stretchr/testify/mock"
)

type EdgeMock struct {
	Source string
	Target string
}

func CreateGraphMock(
	vertices map[string]Node,
	startNode Node,
	edges ...EdgeMock,
) *Graph {
	g := graph.New(func(n Node) string { return n.Path() }, graph.Directed(), graph.Tree(), graph.PreventCycles())

	for _, v := range vertices {
		_ = g.AddVertex(v)
	}

	for _, e := range edges {
		_ = g.AddEdge(e.Source, e.Target)
	}

	return &Graph{
		Graph:     g,
		StartNode: startNode,
	}
}

type NodeMock struct {
	mock.Mock
	tainted  bool
	targeted bool
	oldHash  string
}

func (n *NodeMock) SetOldHash(hash string) {
	n.oldHash = hash
}

func (n *NodeMock) GetOldHash() string {
	return n.oldHash
}

func (n *NodeMock) Path() string {
	args := n.Called()
	return args.String(0)
}

func (n *NodeMock) Identifier() string {
	args := n.Called()
	return args.String(0)
}

func (n *NodeMock) Type() Type {
	args := n.Called()
	return args.Get(0).(Type)
}

func (n *NodeMock) Ancestor() Node {
	//TODO implement me
	panic("implement me")
}

func (n *NodeMock) Parents() ([]Node, error) {
	args := n.Called()
	return args.Get(0).([]Node), args.Error(1)
}

func (n *NodeMock) Independent() bool {
	//TODO implement me
	panic("implement me")
}

func (n *NodeMock) Tainted() bool {
	return n.tainted
}

func (n *NodeMock) Targeted() bool {
	args := n.Called()
	return args.Bool(0)
}

func (n *NodeMock) Hash() (string, error) {
	args := n.Called()
	return args.String(0), args.Error(1)
}

func (n *NodeMock) SetTainted(tainted bool) {
	n.tainted = tainted
}

func (n *NodeMock) SetTargeted(targeted bool) {
	n.targeted = targeted
}

func (n *NodeMock) HasChanges() (bool, error) {
	args := n.Called()
	return args.Bool(0), args.Error(1)
}

func (n *NodeMock) resetGraph(g graph.Graph[string, Node]) {
	_ = n.Called(g)

	return
}
