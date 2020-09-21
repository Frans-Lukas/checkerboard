package cellmanager

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell"
	"github.com/Frans-Lukas/checkerboard/pkg/generated"
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
	cellManager.AppendCell(cell.Cell{CellId: in.CellId})
	return &generated.CellStatusReply{WasPerformed: true}, nil
}

func (cellManager *CellManager) AppendCell(cell cell.Cell) {
	*cellManager.Cells = append(*cellManager.Cells, cell)
}
