package dependency

import (
	"context"
	"github.com/mach-composer/mach-composer-cli/internal/config"
)

const (
	ProjectType       Type = "project"
	SiteType          Type = "site"
	SiteComponentType Type = "site-component"
)

type Type string

type Node interface {
	Path() string
	Identifier() string
	Type() Type
	Parent() Node
	Independent() bool
	HasConfigChanges(ctx context.Context) (bool, error)
	Tainted() bool
	SetTainted(tainted bool)
	Hash() (string, error)
}

type node struct {
	path           string
	identifier     string
	typ            Type
	parent         Node
	deploymentType config.DeploymentType
	tainted        bool
}

func (n *node) SetTainted(tainted bool) {
	n.tainted = tainted
}

func (n *node) Tainted() bool {
	return n.tainted
}

func (n *node) Path() string {
	return n.path
}

func (n *node) Identifier() string {
	return n.identifier
}

func (n *node) Type() Type {
	return n.typ
}

func (n *node) Parent() Node {
	return n.parent
}

func (n *node) Independent() bool {
	// Projects and sites are always independent elements
	if n.typ == ProjectType || n.typ == SiteType {
		return true
	}

	//If deployment type is set to site-component and node is a site-component we consider it independent also
	if n.typ == SiteComponentType && n.deploymentType == config.DeploymentSiteComponent {
		return true
	}

	return false
}
