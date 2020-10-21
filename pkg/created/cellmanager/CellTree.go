package cellmanager

import (
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	"time"
)

type CellTreeNode struct {
	Parent       *CellTreeNode
	Children     *[4]*CellTreeNode
	count        *int
	CreationTime *time.Duration
	*objects.Cell
}

func CreateCellTree(cell *objects.Cell) *CellTreeNode {
	newNode := CreateCellTreeNode(cell)
	newNode.Parent = nil
	return newNode
}

func CreateCellTreeNode(cell *objects.Cell) *CellTreeNode {
	var children [4]*CellTreeNode

	count := 0

	timeNow := timeNowInSeconds()

	return &CellTreeNode{count: &count, Cell: cell, Children: &children, Parent: nil, CreationTime: &timeNow}
}

func (node *CellTreeNode) CreateChild(cell *objects.Cell) *CellTreeNode {
	newNode := CreateCellTreeNode(cell)
	newNode.Parent = node
	return newNode
}

func (node *CellTreeNode) isRoot() bool {
	return node.Parent == nil
}

func (node *CellTreeNode) isLeaf() bool {
	return (*node.Children)[0] == nil
}

func (node *CellTreeNode) addChildren(c1 *objects.Cell, c2 *objects.Cell, c3 *objects.Cell, c4 *objects.Cell) {
	node.resetTimer()
	node.Children[0] = node.CreateChild(c1)
	node.Children[1] = node.CreateChild(c2)
	node.Children[2] = node.CreateChild(c3)
	node.Children[3] = node.CreateChild(c4)
}

func (node *CellTreeNode) IncrementCount() {
	node.changeCount(1)
}

func (node *CellTreeNode) DecrementCount() {
	node.changeCount(-1)
}

func (node *CellTreeNode) changeCount(count int) {
	*node.count += count
	if !node.isRoot() {
		node.Parent.changeCount(count)
	} else {
		println("Incrementing root node")
	}
}

func (node *CellTreeNode) findNode(CellId string) *CellTreeNode {
	if node.CellId == CellId {
		return node
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

func (node *CellTreeNode) findCollidingCell(position *cellmanager.Position) *CellTreeNode {
	if node.Cell.CollidesWith(position) && node.isLeaf() {
		return node
	} else if node.Cell.CollidesWith(position) {
		for _, child := range node.Children {
			result := child.findCollidingCell(position)

			if result != nil {
				return result
			}
		}
	}
	return nil
}

func (node *CellTreeNode) printTree(level int) {
	for i := 0; i < level; i++ {
		print("\t")

	}
	println("id: ", node.CellId, ", x: ", node.PosX, ", y: ", node.PosY, ", w: ", node.Width, ", h: ", node.Height, ", count: ", *node.count, ", len(players): ", len(node.Players))

	for _, node := range node.Children {
		if node != nil {
			node.printTree(level + 1)
		}
	}
}

func (node *CellTreeNode) killChildren() {
	node.Children[0] = nil
	node.Children[1] = nil
	node.Children[2] = nil
	node.Children[3] = nil
}

func (node *CellTreeNode) findMergableCell() (bool, *CellTreeNode) {

	if node.isLeaf() {
		return false, nil
	}

	println("Checking to merge cell: ", node.CellId, ": ", )

	if node.shouldMerge() {
		return true, node
	}

	for _, child := range node.Children {
		shouldMerge, node := child.findMergableCell()
		if shouldMerge {
			return shouldMerge, node
		}
	}

	// will never happen
	return false, nil
}
func (node *CellTreeNode) findSplittableCell() (bool, *CellTreeNode) {
	if node.shouldSplit() && node.isLeaf() {
		return true, node
	} else if node.isLeaf() {
		return false, nil
	}

	for _, child := range node.Children {
		shouldSplit, node := child.findSplittableCell()
		if shouldSplit {
			return shouldSplit, node
		}
	}

	// will never happen
	return false, nil
}

func timeNowInSeconds() time.Duration {
	return time.Duration(time.Now().UnixNano()) / time.Second
}

func (node *CellTreeNode) resetTimer() {
	println("Resetting timer for cell ", node.CellId)
	timeNow := timeNowInSeconds()
	node.CreationTime = &timeNow
	println("time diff after restet ", timeNowInSeconds()-*node.CreationTime)
	if !node.isRoot() {
		node.Parent.resetTimer()
	}
}

func (node *CellTreeNode) shouldMerge() (bool) {
	println(timeNowInSeconds() - *node.CreationTime)
	return *node.count <= constants.MergeCellRequirement && (timeNowInSeconds()-*node.CreationTime) > constants.MergeAgeRequirement
}
func (node *CellTreeNode) shouldSplit() (bool) {
	return *node.count >= constants.SplitCellRequirement
}

func (node *CellTreeNode) retrieveChildren(cell *objects.Cell) {
	for _, player := range node.Cell.Players {
		if cell.ContainsPlayer(player) {
			continue
		}
		cell.AppendPlayer(player)
	}

	if node.isLeaf() {
		return
	}

	for _, child := range node.Children {
		child.retrieveChildren(cell)
	}
}

// Leave cell decrement
// Join cell increment