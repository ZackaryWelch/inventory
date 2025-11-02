//go:build !js || !wasm

package app

import (
	"fmt"

	"cogentcore.org/core/core"
	"github.com/nishiki/frontend/pkg/types"
)

// ContainerNode represents a node in the container hierarchy tree
type ContainerNode struct {
	Container *types.Container
	Children  []*ContainerNode
	Parent    *ContainerNode
	Expanded  bool
}

// BuildContainerHierarchy builds a hierarchical tree from a flat list of containers (desktop stub)
func (app *App) BuildContainerHierarchy(containers []types.Container) []*ContainerNode {
	// Desktop implementation stub
	return make([]*ContainerNode, 0)
}

// RenderContainerTree renders a hierarchical tree of containers (desktop stub)
func (app *App) RenderContainerTree(parent core.Widget, nodes []*ContainerNode, level int) {
	// Desktop implementation stub
}

// RenderContainerTreeNode renders a single node in the container tree (desktop stub)
func (app *App) RenderContainerTreeNode(parent core.Widget, node *ContainerNode, level int) {
	// Desktop implementation stub
}

// RenderCapacityIndicator renders a visual capacity indicator (desktop stub)
func (app *App) RenderCapacityIndicator(parent core.Widget, container *types.Container) {
	// Desktop implementation stub
}

// showContainerDetail shows detailed view of a container (desktop stub)
func (app *App) showContainerDetail(container *types.Container) {
	core.MessageSnackbar(app, fmt.Sprintf("Container detail for: %s", container.Name))
}

// renderStat renders a stat item (desktop stub)
func (app *App) renderStat(parent core.Widget, label, value string) {
	// Desktop implementation stub
}

// RenderFullCapacityBar renders a detailed capacity bar with labels (desktop stub)
func (app *App) RenderFullCapacityBar(parent core.Widget, container *types.Container) {
	// Desktop implementation stub
}

// showContainerActions shows a menu of actions for a container (desktop stub)
func (app *App) showContainerActions(container *types.Container) {
	core.MessageSnackbar(app, fmt.Sprintf("Actions for %s", container.Name))
}
