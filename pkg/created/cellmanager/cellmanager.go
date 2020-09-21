package cellmanager

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell"
	"github.com/Frans-Lukas/checkerboard/pkg/generated"
	"sort"
)

type CellManager struct {
	generated.CellManagerServer
	Cells *[]cell.Cell
}

func NewCellManager() CellManager {
	cells := make([]cell.Cell, 0)
	return CellManager{Cells: &cells}
}

func (cellManager *CellManager) CreateCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellStatusReply, error) {
	cellManager.AppendCell(cell.Cell{CellId: in.CellId, Players: make([]cell.Player, 0)})
	return &generated.CellStatusReply{WasPerformed: true}, nil
}

func (cellManager *CellManager) DeleteCell(
	ctx context.Context, in *generated.CellRequest,
) (*generated.CellStatusReply, error) {
	length := len(*cellManager.Cells)
	i := sort.Search(length, func(i int) bool { return in.CellId == (*cellManager.Cells)[i].CellId })
	if i != length {
		(*cellManager.Cells)[i] = (*cellManager.Cells)[length - 1]
		*cellManager.Cells = (*cellManager.Cells)[:length - 1]
		return &generated.CellStatusReply{WasPerformed: true}, nil
	} else {
		return &generated.CellStatusReply{WasPerformed: false}, nil
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
	return &generated.CellMasterReply{}, nil
}

func (cellManager *CellManager) UnregisterCellMaster(
	ctx context.Context, in *generated.CellMasterRequest,
) (*generated.CellMasterStatusReply, error) {
	return &generated.CellMasterStatusReply{}, nil
}

func (cellManager *CellManager) PlayerLeftCell(
	ctx context.Context, in *generated.PlayerLeftCellRequest,
) (*generated.PlayerStatusReply, error) {

	return &generated.PlayerStatusReply{}, nil
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
	return &generated.CellLockStatusReply{}, nil
}

func (cellManager *CellManager) UnlockCells(
	ctx context.Context, in *generated.LockCellsRequest,
) (*generated.CellLockStatusReply, error) {
	return &generated.CellLockStatusReply{}, nil
}

func (cellManager *CellManager) AppendCell(cell cell.Cell) {
	*cellManager.Cells = append(*cellManager.Cells, cell)
}
