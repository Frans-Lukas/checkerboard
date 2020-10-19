package cellmanager

import (
	"context"
	"errors"
	"fmt"
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

// should only be called once!!!!!!!!!!!!!!!
func (cellManager *CellManager) CreateCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellStatusReply, error) {

	cellToAdd := objects.Cell{CellId: in.CellId, Players: make([]objects.Client, 0)}

	cellManager.AppendCell(cellToAdd)

	cellManager.CellTree = CreateCellTree(cellToAdd)

	return &generated.CellStatusReply{WasPerformed: true}, nil
}

func (cellManager *CellManager) SetWorldSize(
	ctx context.Context, in *generated.WorldSize,
) (*generated.TransactionSucceeded, error) {
	cellManager.WorldWidth = in.Width
	cellManager.WorldHeight = in.Height

	if len(*cellManager.Cells) > 0 {
		return &generated.TransactionSucceeded{Succeeded: false}, nil
	}

	cellManager.AppendCell(objects.Cell{
		CellId:  "initialCell",
		Players: make([]objects.Client, 0),
		PosY:    0,
		PosX:    0,
		Width:   in.Width,
		Height:  in.Height,
	})

	return &generated.TransactionSucceeded{Succeeded: true}, nil
}

func (cellManager *CellManager) AddPlayerToCell(
	ctx context.Context, in *generated.PlayerInCellRequest,
) (*generated.TransactionSucceeded, error) {
	for index, cell := range *cellManager.Cells {
		if cell.CellId == in.CellId {
			(*cellManager.Cells)[index].AppendPlayer(
				objects.Client{
					Ip:         in.Ip,
					Port:       in.Port,
					TrustLevel: 0,
				},
			)
			return &generated.TransactionSucceeded{Succeeded: true}, nil
		}
	}
	return &generated.TransactionSucceeded{Succeeded: false}, errors.New("Invalid cellID: " + in.CellId)
}

func (cellManager *CellManager) AddPlayerToCellWithPositions(
	ctx context.Context, in *generated.PlayerInCellRequestWithPositions,
) (*generated.TransactionSucceeded, error) {
	for index, cell := range *cellManager.Cells {
		if cell.CollidesWith(&generated.Position{PosY: in.PosY, PosX: in.PosX}) {

			cellManager.CellTree.IncrementCount(cell.CellId)

			(*cellManager.Cells)[index].AppendPlayer(
				objects.Client{
					Ip:         in.Ip,
					Port:       in.Port,
					TrustLevel: 0,
				},
			)
			return &generated.TransactionSucceeded{Succeeded: true}, nil

		}
	}
	return &generated.TransactionSucceeded{Succeeded: false}, errors.New("Invalid position: x: " + strconv.FormatInt(in.PosX, 10) + ", y: " + strconv.FormatInt(in.PosY, 10))
}

func (cellManager *CellManager) RequestCellMasterWithPositions(
	ctx context.Context, in *generated.Position,
) (*generated.CellMasterReply, error) {
	for cellIndex, cell := range *cellManager.Cells {
		if cell.CollidesWith(in) {
			cm, err := cellManager.selectCellMaster(cell, cellIndex)
			if err != nil {
				println("request cell master: no player")
				return &generated.CellMasterReply{Ip: "no player", Port: - 1}, errors.New("no player in cell")
			}
			go func() {
				NotifyOfCellMastership(*cm, cell)
			}()
			//helpers.DebugPrint(requestCMWithPosPrint, fmt.Sprintf("returning cm with port: $d", cm.Port))
			println("request cell master: found cell master ", cm.Ip, ":", cm.Port)
			return &generated.CellMasterReply{Ip: cm.Ip, Port: cm.Port}, nil
		}
	}
	println("request cell master: invalid position ")
	return &generated.CellMasterReply{Ip: "INVALID POSITION", Port: -1}, errors.New("Invalid position: x: " + strconv.FormatInt(in.PosX, 10) + ", y: " + strconv.FormatInt(in.PosY, 10))
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

func (cellManager *CellManager) DeleteCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellStatusReply, error) {
	length := len(*cellManager.Cells)

	index := -1
	for i, storedCell := range *cellManager.Cells {
		if in.CellId == storedCell.CellId {
			index = i
		}
	}

	if index == -1 {
		return &generated.CellStatusReply{WasPerformed: false}, nil
	} else {
		(*cellManager.Cells)[index] = (*cellManager.Cells)[length-1]
		*cellManager.Cells = (*cellManager.Cells)[:length-1]
		return &generated.CellStatusReply{WasPerformed: true}, nil
	}
}

func (cellManager *CellManager) ListCells(
	ctx context.Context, in *generated.ListCellsRequest,
) (*generated.ListCellsReply, error) {
	cellIds := make([]string, len(*cellManager.Cells))
	for index, cell := range *cellManager.Cells {
		cellIds[index] = cell.CellId
	}
	cells := generated.ListCellsReply{CellId: cellIds}
	return &cells, nil
}

func (cellManager *CellManager) ListPlayersInCell(
	ctx context.Context, in *generated.ListPlayersRequest,
) (*generated.PlayersReply, error) {
	playerIps := make([]string, 0)
	playerPorts := make([]int32, 0)
	for _, cell := range *cellManager.Cells {
		if cell.CellId == in.CellId {
			for _, player := range cell.Players {
				playerIps = append(playerIps, player.Ip)
				playerPorts = append(playerPorts, player.Port)
			}
		}
	}
	return &generated.PlayersReply{Port: playerPorts, Ip: playerIps}, nil
}

func (cellManager *CellManager) RequestCellMaster(
	ctx context.Context, in *generated.CellMasterRequest,
) (*generated.CellMasterReply, error) {

	for cellIndex, cell := range *cellManager.Cells {
		if in.CellId == cell.CellId {
			return cellManager.selectCellMaster(cell, cellIndex)
		}
	}

	return &generated.CellMasterReply{}, nil
}

func (cellManager *CellManager) selectCellMaster(cell objects.Cell, cellIndex int) (*generated.CellMasterReply, error) {

	if cell.CellMaster == nil {
		cmIndex := cell.SelectNewCellMaster()

		if cmIndex == -1 {
			return &generated.CellMasterReply{Ip: "", Port: -1}, errors.New("empty cell requested a cell master")
		}

		newCM := cell.Players[cmIndex]

		(*cellManager.Cells)[cellIndex].CellMaster = &newCM

		return &generated.CellMasterReply{Ip: newCM.Ip, Port: newCM.Port}, nil
	} else {
		return &generated.CellMasterReply{Ip: cell.CellMaster.Ip, Port: cell.CellMaster.Port}, nil
	}

}

func (cellManager *CellManager) UnregisterCellMaster(
	ctx context.Context, in *generated.CellMasterRequest,
) (*generated.CellMasterStatusReply, error) {
	success := false
	for index, cell := range *cellManager.Cells {
		if cell.CellId == in.CellId {
			(*cellManager.Cells)[index].CellMaster = nil
			success = true
		}
	}
	return &generated.CellMasterStatusReply{WasUnregistered: success}, nil
}

func (cellManager *CellManager) PlayerLeftCell(
	ctx context.Context, in *generated.PlayerInCellRequest,
) (*generated.PlayerStatusReply, error) {
	for _, cellToLeave := range *cellManager.Cells {
		if cellToLeave.CellId == in.CellId {
			cellManager.CellTree.DecrementCount(cellToLeave.CellId)
			cellToLeave.DeletePlayer(objects.Client{Port: in.Port, Ip: in.Ip})
		}
	}
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

	var indexes []int

	for _, cellId := range in.CellId {
		cellIsLockable := false
		for i, storedCell := range *cellManager.Cells {
			if cellId == storedCell.CellId {
				if storedCell.Locked {
					break
				}
				indexes = append(indexes, i)
				cellIsLockable = true
				break
			}
		}
		if !cellIsLockable {
			return &generated.CellLockStatusReply{Locked: false, Lockee: "TODO"}, nil
		}
	}

	for _, j := range indexes {
		(*cellManager.Cells)[j].Locked = true
		(*cellManager.Cells)[j].Lockee = in.SenderCellId
	}

	return &generated.CellLockStatusReply{Locked: true, Lockee: "TODO"}, nil
}

func (cellManager *CellManager) UnlockCells(
	ctx context.Context, in *generated.LockCellsRequest,
) (*generated.CellLockStatusReply, error) {

	var indexes []int

	for _, cellId := range in.CellId {
		cellIsUnlockable := false
		for i, storedCell := range *cellManager.Cells {
			if cellId == storedCell.CellId {
				if !storedCell.Locked || storedCell.Lockee != in.SenderCellId {
					break
				}
				indexes = append(indexes, i)
				cellIsUnlockable = true
				break
			}
		}
		if !cellIsUnlockable {
			return &generated.CellLockStatusReply{Locked: true, Lockee: "TODO"}, nil
		}
	}

	for _, j := range indexes {
		(*cellManager.Cells)[j].Locked = false
		(*cellManager.Cells)[j].Lockee = ""
	}

	return &generated.CellLockStatusReply{Locked: false, Lockee: "TODO"}, nil
}

func (cellManager *CellManager) DivideCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellChangeStatusReply, error) {
	cellIndex := FindCell(*cellManager.Cells, in.CellId)
	if cellIndex != len(*cellManager.Cells) {
		cell := (*cellManager.Cells)[cellIndex]

		if cell.Locked {
			return &generated.CellChangeStatusReply{Succeeded: false}, errors.New("cell is locked")
		}
		newWidth := int64(UpDiv(int(cell.Width), 2))
		newHeight := int64(UpDiv(int(cell.Height), 2))

		nodeToAddChildrenTo := cellManager.CellTree.findNode(in.CellId)

		cell1 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: cell.PosX, PosY: cell.PosY, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
		cellManager.CellIDNumber++
		cell2 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: cell.PosX, PosY: cell.PosY + cell.Height/2, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
		cellManager.CellIDNumber++
		cell3 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: cell.PosX + cell.Width/2, PosY: cell.PosY, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
		cellManager.CellIDNumber++
		cell4 := objects.Cell{CellId: strconv.Itoa(int(cellManager.CellIDNumber)), PosX: cell.PosX + cell.Width/2, PosY: cell.PosY + cell.Height/2, Width: newWidth, Height: newHeight, Players: make([]objects.Client, 0)}
		cellManager.CellIDNumber++
		cellIndex := FindCell(*cellManager.Cells, in.CellId)

		nodeToAddChildrenTo.addChildren(cell1, cell2, cell3, cell4)

		(*cellManager.Cells)[cellIndex] = cell1
		cellManager.AppendCell(cell2)
		cellManager.AppendCell(cell3)
		cellManager.AppendCell(cell4)

		return &generated.CellChangeStatusReply{Succeeded: true}, nil
	} else {
		return &generated.CellChangeStatusReply{Succeeded: false}, errors.New("cellId does not match an existing cell")
	}
}

func (cellManager *CellManager) AppendCell(cell objects.Cell) {
	*cellManager.Cells = append(*cellManager.Cells, cell)
}

func (cellManager *CellManager) TryToMergeCell(cell1 objects.Cell) bool {
	cell1Corner1X := int64(0)
	cell1Corner1Y := int64(0)
	cell1Corner2X := int64(0)
	cell1Corner2Y := int64(0)

	cell2Corner1X := int64(0)
	cell2Corner1Y := int64(0)
	cell2Corner2X := int64(0)
	cell2Corner2Y := int64(0)

	// check left, up, right, down
	for index, cell2 := range *cellManager.Cells {
		// check if locked
		if !cell2.Locked && cell2.CellId != cell1.CellId {
			for direction := 0; direction < 4; direction++ {
				switch direction {
				case 0: // left
					cell1Corner1X = cell1.PosX
					cell1Corner1Y = cell1.PosY
					cell1Corner2X = cell1.PosX
					cell1Corner2Y = cell1.PosY + cell1.Height
					cell2Corner1X = cell2.PosX + cell2.Width
					cell2Corner1Y = cell2.PosY
					cell2Corner2X = cell2Corner1X
					cell2Corner2Y = cell2.PosY + cell2.Height
					break
				case 1: // up
					cell1Corner1X = cell1.PosX
					cell1Corner1Y = cell1.PosY
					cell1Corner2X = cell1.PosX + cell1.Width
					cell1Corner2Y = cell1.PosY
					cell2Corner1X = cell2.PosX
					cell2Corner1Y = cell2.PosY + cell2.Height
					cell2Corner2X = cell2.PosX + cell2.Width
					cell2Corner2Y = cell2.PosY + cell2.Height
					break
				case 2: // right
					cell1Corner1X = cell1.PosX + cell1.Width
					cell1Corner1Y = cell1.PosY
					cell1Corner2X = cell1.PosX + cell1.Width
					cell1Corner2Y = cell1.PosY + cell1.Height
					cell2Corner1X = cell2.PosX
					cell2Corner1Y = cell2.PosY
					cell2Corner2X = cell2.PosX
					cell2Corner2Y = cell2.PosY + cell2.Height
					break
				case 3: // down
					cell1Corner1X = cell1.PosX
					cell1Corner1Y = cell1.PosY + cell1.Height
					cell1Corner2X = cell1.PosX + cell1.Width
					cell1Corner2Y = cell1.PosY + cell1.Height
					cell2Corner1X = cell2.PosX
					cell2Corner1Y = cell2.PosY
					cell2Corner2X = cell2.PosX + cell2.Width
					cell2Corner2Y = cell2.PosY
					break
				}

				// check if cell2 connects
				if cell1Corner1X == cell2Corner1X && cell1Corner1Y == cell2Corner1Y && cell1Corner2X == cell2Corner2X && cell1Corner2Y == cell2Corner2Y {
					// merge
					tmpCell := objects.NewCellFromCells(strconv.Itoa(int(cellManager.CellIDNumber)), cell1, cell2)
					cell1.PosX = tmpCell.PosX
					cell1.PosY = tmpCell.PosY
					cell1.Width = tmpCell.Width
					cell1.Height = tmpCell.Height

					// replace cell1 and remove cell2
					cellIndex := FindCell(*cellManager.Cells, cell1.CellId)
					(*cellManager.Cells)[cellIndex] = cell1
					(*cellManager.Cells)[index] = (*cellManager.Cells)[len(*cellManager.Cells)-1]
					*cellManager.Cells = (*cellManager.Cells)[:len(*cellManager.Cells)-1]

					// inform all cell members of removal from cell
					//for _, client := range cell1.Players {
					//	cellManager.InformClientOfCellMasterChange(client)
					//}
					if cell1.CellMaster != nil {
						cellManager.InformCellMasterOfCellChange(*cell1.CellMaster, cell1)
					}
					for _, client := range cell2.Players {
						cellManager.InformClientOfCellMasterChange(client)
					}
					return true
				}
			}
		}
	}
	return false
}

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

func FindCell(cells []objects.Cell, cellId string) int {
	for i, n := range cells {
		if cellId == n.CellId {
			return i
		}
	}
	return len(cells)
}
