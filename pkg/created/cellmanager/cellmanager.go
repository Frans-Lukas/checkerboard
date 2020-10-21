package cellmanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	objects2 "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"time"
)

const requestCMWithPosPrint = true

type CellManager struct {
	generated.CellManagerServer
	WorldWidth   int64
	WorldHeight  int64
	Cells        *[]objects.Cell
	CellIDNumber int64
	CellTree     *CellTreeNode
}

func NewCellManager() CellManager {
	cells := make([]objects.Cell, 0)
	return CellManager{Cells: &cells, CellIDNumber: 0}
}

func (cellManager *CellManager) SetWorldSize(
	ctx context.Context, in *generated.WorldSize,
) (*generated.TransactionSucceeded, error) {
	cellManager.WorldWidth = in.Width
	cellManager.WorldHeight = in.Height

	if cellManager.CellTree != nil {
		return &generated.TransactionSucceeded{Succeeded: false}, nil
	}

	cellManager.CellTree = CreateCellTree(&objects.Cell{
		CellId:  "initialCell",
		Players: make([]objects.Client, 0),
		PosY:    0,
		PosX:    0,
		Width:   in.Width,
		Height:  in.Height,
	})

	return &generated.TransactionSucceeded{Succeeded: true}, nil
}

//
//func (cellManager *CellManager) AddPlayerToCell(
//	ctx context.Context, in *generated.PlayerInCellRequest,
//) (*generated.TransactionSucceeded, error) {
//
//
//
//	for index, cell := range *cellManager.Cells {
//		if cell.CellId == in.CellId {
//			(*cellManager.Cells)[index].AppendPlayer(
//				objects.Client{
//					Ip:         in.Ip,
//					Port:       in.Port,
//					TrustLevel: 0,
//				},
//			)
//			return &generated.TransactionSucceeded{Succeeded: true}, nil
//		}
//	}
//	return &generated.TransactionSucceeded{Succeeded: false}, errors.New("Invalid cellID: " + in.CellId)
//}

func (cellManager *CellManager) AddPlayerToCellWithPositions(
	ctx context.Context, in *generated.PlayerInCellRequestWithPositions,
) (*generated.TransactionSucceeded, error) {

	collidingCell := cellManager.CellTree.findCollidingCell(&generated.Position{PosY: in.PosY, PosX: in.PosX})

	if collidingCell == nil {
		println("request cell master: invalid position ")
		return &generated.TransactionSucceeded{Succeeded: false}, errors.New("Invalid position: x: " + strconv.FormatInt(in.PosX, 10) + ", y: " + strconv.FormatInt(in.PosY, 10))
	}

	println("Adding player to cellID: ", collidingCell.CellId)

	playerToAdd := objects.Client{
		Ip:         in.Ip,
		Port:       in.Port,
		TrustLevel: 0,
	}

	if collidingCell.ContainsPlayer(playerToAdd) {
		return &generated.TransactionSucceeded{Succeeded: true}, nil
	}
	collidingCell.IncrementCount()
	collidingCell.AppendPlayer(playerToAdd)
	return &generated.TransactionSucceeded{Succeeded: true}, nil
}

func (cellManager *CellManager) RequestCellMasterWithPositions(
	ctx context.Context, in *generated.Position,
) (*generated.CellMasterReply, error) {

	collidingCell := cellManager.CellTree.findCollidingCell(in)

	if collidingCell == nil {
		println("request cell master: invalid position ")
		return &generated.CellMasterReply{Ip: "INVALID POSITION", Port: -1}, errors.New("Invalid position: x: " + strconv.FormatInt(in.PosX, 10) + ", y: " + strconv.FormatInt(in.PosY, 10))
	}

	cm, err := cellManager.selectCellMaster(*collidingCell.Cell)
	if err != nil {
		println("request cell master: no player")
		return &generated.CellMasterReply{Ip: "no player", Port: - 1}, errors.New("no player in cell")
	}
	go func() {
		NotifyOfCellMastership(*cm, *collidingCell.Cell)
	}()

	println("request cell master: found cell master ", cm.Ip, ":", cm.Port)
	return &generated.CellMasterReply{Ip: cm.Ip, Port: cm.Port}, nil
}

func NotifyOfCellMastership(reply generated.CellMasterReply, cell objects.Cell) {
	address := fmt.Sprintf(reply.Ip + ":" + strconv.Itoa(int(reply.Port)))
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := objects2.NewPlayerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	newCell := cell.ToGeneratedCell()
	c.ReceiveCellMastership(ctx, &objects2.CellList{Cells: []*objects2.Cell{&newCell}})
}

//func (cellManager *CellManager) DeleteCell(
//	ctx context.Context, in *generated.CellRequest,
//) (*generated.CellStatusReply, error) {
//	length := len(*cellManager.Cells)
//
//	index := -1
//	for i, storedCell := range *cellManager.Cells {
//		if in.CellId == storedCell.CellId {
//			index = i
//		}
//	}
//
//	if index == -1 {
//		return &generated.CellStatusReply{WasPerformed: false}, nil
//	} else {
//		(*cellManager.Cells)[index] = (*cellManager.Cells)[length-1]
//		*cellManager.Cells = (*cellManager.Cells)[:length-1]
//		return &generated.CellStatusReply{WasPerformed: true}, nil
//	}
//}

//func (cellManager *CellManager) ListCells(
//	ctx context.Context, in *generated.ListCellsRequest,
//) (*generated.ListCellsReply, error) {
//	cellIds := make([]string, len(*cellManager.Cells))
//	for index, cell := range *cellManager.Cells {
//		cellIds[index] = cell.CellId
//	}
//	cells := generated.ListCellsReply{CellId: cellIds}
//	return &cells, nil
//}
//
//func (cellManager *CellManager) ListPlayersInCell(
//	ctx context.Context, in *generated.ListPlayersRequest,
//) (*generated.PlayersReply, error) {
//	playerIps := make([]string, 0)
//	playerPorts := make([]int32, 0)
//
//
//
//	for _, cell := range *cellManager.Cells {
//		if cell.CellId == in.CellId {
//			for _, player := range cell.Players {
//				playerIps = append(playerIps, player.Ip)
//				playerPorts = append(playerPorts, player.Port)
//			}
//		}
//	}
//	return &generated.PlayersReply{Port: playerPorts, Ip: playerIps}, nil
//}

func (cellManager *CellManager) RequestCellMaster(
	ctx context.Context, in *generated.CellMasterRequest,
) (*generated.CellMasterReply, error) {

	node := cellManager.CellTree.findNode(in.CellId)

	if node == nil {
		return &generated.CellMasterReply{}, errors.New("invalid cell")
	}

	return cellManager.selectCellMaster(*node.Cell)
}

func (cellManager *CellManager) selectCellMaster(cell objects.Cell) (*generated.CellMasterReply, error) {

	if cell.CellMaster == nil {
		cmIndex := cell.SelectNewCellMaster()

		if cmIndex == -1 {
			return &generated.CellMasterReply{Ip: "", Port: -1}, errors.New("empty cell requested a cell master")
		}

		newCM := cell.Players[cmIndex]

		cellToAddTo := cellManager.CellTree.findNode(cell.CellId)

		if cellToAddTo == nil {
			return &generated.CellMasterReply{Ip: "", Port: -1}, errors.New("empty cell requested a cell master")
		}

		cellToAddTo.CellMaster = &newCM

		return &generated.CellMasterReply{Ip: newCM.Ip, Port: newCM.Port}, nil
	} else {
		return &generated.CellMasterReply{Ip: cell.CellMaster.Ip, Port: cell.CellMaster.Port}, nil
	}

}

func (cellManager *CellManager) UnregisterCellMaster(
	ctx context.Context, in *generated.CellMasterRequest,
) (*generated.CellMasterStatusReply, error) {

	cellToUnregister := cellManager.CellTree.findNode(in.CellId)

	if cellToUnregister == nil {
		return &generated.CellMasterStatusReply{WasUnregistered: false}, errors.New("invalid cell to unregister from")
	}

	cellToUnregister.Cell.CellMaster = nil
	return &generated.CellMasterStatusReply{WasUnregistered: true}, nil
}

func (cellManager *CellManager) PlayerLeftCell(
	ctx context.Context, in *generated.PlayerInCellRequest,
) (*generated.PlayerStatusReply, error) {

	cellToLeave := cellManager.CellTree.findNode(in.CellId)

	if cellToLeave == nil {
		return &generated.PlayerStatusReply{PlayerLeft: false}, errors.New("invalid cell to delete from")
	}

	cellToLeave.DecrementCount()

	cellToLeave.Cell.DeletePlayer(objects.Client{Port: in.Port, Ip: in.Ip})
	return &generated.PlayerStatusReply{PlayerLeft: true}, nil
}

func (cellManager *CellManager) RequestCellNeighbours(
	ctx context.Context, in *generated.CellNeighbourRequest,
) (*generated.CellNeighboursReply, error) {
	return &generated.CellNeighboursReply{}, nil
}

func (cellManager *CellManager) RequestCellSizeChange(
	ctx context.Context, in *generated.CellChangeSizeRequest,
) (*generated.CellChangeStatusReply, error) {
	return &generated.CellChangeStatusReply{}, nil
}

func (cellManager *CellManager) LockCells(
	ctx context.Context, in *generated.LockCellsRequest,
) (*generated.CellLockStatusReply, error) {

	var cellsToLock []*CellTreeNode

	for _, cellId := range in.CellId {
		storedCell := cellManager.CellTree.findNode(cellId)
		if storedCell == nil {
			return &generated.CellLockStatusReply{Locked: false, Lockee: "TODO"}, errors.New("invalid cellid given")
		}

		if storedCell.Locked {
			return &generated.CellLockStatusReply{Locked: true, Lockee: "TODO"}, errors.New("at least one cell is already locked")
		}
		cellsToLock = append(cellsToLock, storedCell)
	}

	for _, treeNode := range cellsToLock {

		treeNode.Cell.Locked = true
		treeNode.Cell.Lockee = in.SenderCellId
	}

	return &generated.CellLockStatusReply{Locked: true, Lockee: "TODO"}, nil
}

func (cellManager *CellManager) UnlockCells(
	ctx context.Context, in *generated.LockCellsRequest,
) (*generated.CellLockStatusReply, error) {

	var cellsToUnlock []*CellTreeNode

	for _, cellId := range in.CellId {
		storedCell := cellManager.CellTree.findNode(cellId)
		if storedCell == nil {
			return &generated.CellLockStatusReply{Locked: false, Lockee: "TODO"}, errors.New("invalid cellid given")
		}

		if !storedCell.Locked || storedCell.Lockee != in.SenderCellId {
			return &generated.CellLockStatusReply{Locked: true, Lockee: "TODO"}, errors.New("at least one cell is already locked")
		}
		cellsToUnlock = append(cellsToUnlock, storedCell)
	}

	for _, j := range cellsToUnlock {
		j.Cell.Locked = false
		j.Cell.Lockee = ""
	}

	return &generated.CellLockStatusReply{Locked: false, Lockee: "TODO"}, nil
}

func (cellManager *CellManager) DivideCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellChangeStatusReply, error) {

	node := cellManager.CellTree.findNode(in.CellId)

	if node == nil {
		return &generated.CellChangeStatusReply{Succeeded: false}, errors.New("cellId does not match an existing cell")
	}

	cell := &node.Cell

	if (*cell).Locked {
		return &generated.CellChangeStatusReply{Succeeded: false}, errors.New("cell is locked")
	}

	newWidth := int64(UpDiv(int((*cell).Width), 2))
	newHeight := int64(UpDiv(int((*cell).Height), 2))

	cell1 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: (*cell).PosX, PosY: (*cell).PosY, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
	cellManager.CellIDNumber++
	cell2 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: (*cell).PosX, PosY: (*cell).PosY + (*cell).Height/2, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
	cellManager.CellIDNumber++
	cell3 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: (*cell).PosX + (*cell).Width/2, PosY: (*cell).PosY, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
	cellManager.CellIDNumber++
	cell4 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: (*cell).PosX + (*cell).Width/2, PosY: (*cell).PosY + (*cell).Height/2, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
	cellManager.CellIDNumber++

	node.changeCount(*node.count * -1)

	node.addChildren(&cell1, &cell2, &cell3, &cell4)

	println("TREEE IS SPLIT, PRINTING: ")
	cellManager.CellTree.printTree()

	return &generated.CellChangeStatusReply{Succeeded: true}, nil
}

func (cellManager *CellManager) AppendCell(cell objects.Cell) {
	*cellManager.Cells = append(*cellManager.Cells, cell)
}

//
//func (cellManager *CellManager) TryToMergeCell(cell1 objects.Cell) bool {
//	cell1Corner1X := int64(0)
//	cell1Corner1Y := int64(0)
//	cell1Corner2X := int64(0)
//	cell1Corner2Y := int64(0)
//
//	cell2Corner1X := int64(0)
//	cell2Corner1Y := int64(0)
//	cell2Corner2X := int64(0)
//	cell2Corner2Y := int64(0)
//
//	// check left, up, right, down
//	for index, cell2 := range *cellManager.Cells {
//		// check if locked
//		if !cell2.Locked && cell2.CellId != cell1.CellId {
//			for direction := 0; direction < 4; direction++ {
//				switch direction {
//				case 0: // left
//					cell1Corner1X = cell1.PosX
//					cell1Corner1Y = cell1.PosY
//					cell1Corner2X = cell1.PosX
//					cell1Corner2Y = cell1.PosY + cell1.Height
//					cell2Corner1X = cell2.PosX + cell2.Width
//					cell2Corner1Y = cell2.PosY
//					cell2Corner2X = cell2Corner1X
//					cell2Corner2Y = cell2.PosY + cell2.Height
//					break
//				case 1: // up
//					cell1Corner1X = cell1.PosX
//					cell1Corner1Y = cell1.PosY
//					cell1Corner2X = cell1.PosX + cell1.Width
//					cell1Corner2Y = cell1.PosY
//					cell2Corner1X = cell2.PosX
//					cell2Corner1Y = cell2.PosY + cell2.Height
//					cell2Corner2X = cell2.PosX + cell2.Width
//					cell2Corner2Y = cell2.PosY + cell2.Height
//					break
//				case 2: // right
//					cell1Corner1X = cell1.PosX + cell1.Width
//					cell1Corner1Y = cell1.PosY
//					cell1Corner2X = cell1.PosX + cell1.Width
//					cell1Corner2Y = cell1.PosY + cell1.Height
//					cell2Corner1X = cell2.PosX
//					cell2Corner1Y = cell2.PosY
//					cell2Corner2X = cell2.PosX
//					cell2Corner2Y = cell2.PosY + cell2.Height
//					break
//				case 3: // down
//					cell1Corner1X = cell1.PosX
//					cell1Corner1Y = cell1.PosY + cell1.Height
//					cell1Corner2X = cell1.PosX + cell1.Width
//					cell1Corner2Y = cell1.PosY + cell1.Height
//					cell2Corner1X = cell2.PosX
//					cell2Corner1Y = cell2.PosY
//					cell2Corner2X = cell2.PosX + cell2.Width
//					cell2Corner2Y = cell2.PosY
//					break
//				}
//
//				// check if cell2 connects
//				if cell1Corner1X == cell2Corner1X && cell1Corner1Y == cell2Corner1Y && cell1Corner2X == cell2Corner2X && cell1Corner2Y == cell2Corner2Y {
//					// merge
//					tmpCell := objects.NewCellFromCells(strconv.Itoa(int(cellManager.CellIDNumber)), cell1, cell2)
//					cell1.PosX = tmpCell.PosX
//					cell1.PosY = tmpCell.PosY
//					cell1.Width = tmpCell.Width
//					cell1.Height = tmpCell.Height
//
//					// replace cell1 and remove cell2
//					cellIndex := FindCell(*cellManager.Cells, cell1.CellId)
//					(*cellManager.Cells)[cellIndex] = cell1
//					(*cellManager.Cells)[index] = (*cellManager.Cells)[len(*cellManager.Cells)-1]
//					*cellManager.Cells = (*cellManager.Cells)[:len(*cellManager.Cells)-1]
//
//					// inform all cell members of removal from cell
//					//for _, client := range cell1.Players {
//					//	cellManager.InformClientOfCellMasterChange(client)
//					//}
//					if cell1.CellMaster != nil {
//						cellManager.InformCellMasterOfCellChange(*cell1.CellMaster, cell1)
//					}
//					for _, client := range cell2.Players {
//						cellManager.InformClientOfCellMasterChange(client)
//					}
//					return true
//				}
//			}
//		}
//	}
//	return false
//}

func (cellManager *CellManager) InformCellMasterOfCellChange(cellMaster objects.Client, cell objects.Cell) {
	address := fmt.Sprintf(cellMaster.Ip + ":" + strconv.Itoa(int(cellMaster.Port)))
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		println("did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := objects2.NewPlayerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	objectCell := objects2.Cell{
		CellId: cell.CellId,
		PosX:   cell.PosX,
		PosY:   cell.PosY,
		Width:  cell.Width,
		Height: cell.Height,
	}

	_, err = c.ReceiveCellMastership(ctx, &objects2.CellList{Cells: []*objects2.Cell{&objectCell}})
	if err != nil {
		println("did not succeed to request receiveCellMasterChip with upated cell: %v", err)
	}
	return
}

func (cellManager *CellManager) InformClientOfCellMasterChange(client objects.Client) {
	address := fmt.Sprintf(client.Ip + ":" + strconv.Itoa(int(client.Port)))
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		println("did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := objects2.NewPlayerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.ChangedCellMaster(ctx, &objects2.ChangedCellMasterRequest{})
	if err != nil {
		println("did not succeed to request changedCellMaster: %v", err)
	}
	return
}

func UpDiv(divident int, divisor int) int {
	return (divident + divisor - 1) / divisor
}

func (cellManager *CellManager) MergeLoop() {
	for {

		if cellManager.CellTree != nil {
			println("root has count: ", *cellManager.CellTree.count)
			shouldMerge, cellToMerge := cellManager.CellTree.findMergableCell()

			if shouldMerge {

				println("Merging cell with player count: ", cellToMerge.count)
				cellManager.performMerge(cellToMerge.CellId)
			}
		}

		time.Sleep(time.Second * constants.SplitCellInterval)
	}
}

func (cellManager *CellManager) performMerge(cellId string) {
	cellToMerge := cellManager.CellTree.findNode(cellId)
	if cellToMerge.isLeaf() {
		return
	}

	cellToMerge.retrieveChildren(cellToMerge.Cell)
	*cellToMerge.count = len(cellToMerge.Players)
	cellToMerge.killChildren()

	for _, player := range cellToMerge.Players {
		cellManager.InformClientOfCellMasterChange(player)
	}
}

//func FindCell(cells []objects.Cell, cellId string) int {
//	for i, n := range cells {
//		if cellId == n.CellId {
//			return i
//		}
//	}
//	return len(cells)
//}
