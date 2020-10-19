package cellmanager

import "github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"

type CellTreeNode struct {
	Parent   *CellTreeNode
	Children [4]*CellTreeNode
	count    int
	objects.Cell
}

func CreateCellTree(cell objects.Cell) *CellTreeNode {
	return &CellTreeNode{count: 0, Cell: cell, Parent: nil}
}

func (node CellTreeNode) CreateChild(cell objects.Cell) *CellTreeNode {
	return &CellTreeNode{count: 0, Cell: cell, Parent: &node}
}

func (node CellTreeNode) isRoot() bool {
	return node.Parent == nil
}

func (node CellTreeNode) isLeaf() bool {
	return node.Children[0] == nil
}

func (node CellTreeNode) addChildren(c1 objects.Cell, c2 objects.Cell, c3 objects.Cell, c4 objects.Cell) {
	node.Children[0] = node.CreateChild(c1)
	node.Children[1] = node.CreateChild(c2)
	node.Children[2] = node.CreateChild(c3)
	node.Children[3] = node.CreateChild(c4)
}

func (node CellTreeNode) IncrementCount(CellID string) {
	node.changeCount(CellID, 1)
}

func (node CellTreeNode) DecrementCount(CellID string) {
	node.changeCount(CellID, -1)
}

func (node CellTreeNode) changeCount(CellID string, count int) {
	nodeToChange := node.findNode(CellID)

	for !nodeToChange.isRoot() {
		nodeToChange.count = nodeToChange.count + count
		nodeToChange = nodeToChange.Parent
	}

	nodeToChange.count = nodeToChange.count + count
}

func (node CellTreeNode) findNode(CellId string) *CellTreeNode {
	if node.CellId == CellId {
		return &node
	}

	if node.isLeaf() {
		return nil
	}

	for _, child := range node.Children {
		result := child.findNode(CellId)

		if result != nil {
			return result
		}
	}

	return nil
}

// Leave cell decrement
// Join cell increment
