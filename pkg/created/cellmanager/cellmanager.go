package cellmanager

import (
	"context"
	"errors"
	created "github.com/Frans-Lukas/checkerboard/pkg/created/cell"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	"strconv"
)

type CellManager struct {
	generated.CellManagerServer
	WorldWidth  int64
	WorldHeight int64
	Cells       *[]created.Cell
}

func NewCellManager() CellManager {
	cells := make([]created.Cell, 0)
	return CellManager{Cells: &cells}
}

func (cellManager *CellManager) CreateCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellStatusReply, error) {
	cellManager.AppendCell(created.Cell{CellId: in.CellId, Players: make([]objects.Client, 0)})
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

	cellManager.AppendCell(created.Cell{
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
	for _, cell := range *cellManager.Cells {
		if cell.CollidesWith(in) {
			return &generated.CellMasterReply{Ip: cell.CellMaster.Ip, Port: cell.CellMaster.Port}, nil
		}
	}
	return &generated.CellMasterReply{Ip: "INVALID POSITION", Port: -1}, errors.New("Invalid position: x: " + strconv.FormatInt(in.PosX, 10) + ", y: " + strconv.FormatInt(in.PosY, 10))
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
	}

	return &generated.CellMasterReply{}, nil
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

func (cellManager *CellManager) AppendCell(cell created.Cell) {
	*cellManager.Cells = append(*cellManager.Cells, cell)
}
