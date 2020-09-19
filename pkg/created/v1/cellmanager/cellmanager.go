package cellmanager

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/pkg/created/v1/cell"
	pb "github.com/Frans-Lukas/checkerboard/pkg/generated/v1"
)

type CellManager struct {
	pb.CellManagerServer
	Cells *[]cell.Cell
}

func (cellManager *CellManager) CreateCell(
	ctx context.Context, in *pb.CellRequest,
) (*pb.CellStatusReply, error) {
	cellManager.AppendCell(cell.Cell{CellId: in.CellId})
	return &pb.CellStatusReply{WasPerformed: true}, nil
}

func (cellManager *CellManager) AppendCell(cell cell.Cell) {
	*cellManager.Cells = append(*cellManager.Cells, cell)
}

func NewCellManager() CellManager {
	cells := make([]cell.Cell, 0)
	return CellManager{Cells: &cells}
}
