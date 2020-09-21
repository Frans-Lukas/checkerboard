package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cellmanager"
	"github.com/Frans-Lukas/checkerboard/pkg/generated"
	"testing"
)

func TestCreateCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := generated.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIfNotNull(err, "could not create cell")
	if (*cm.Cells)[0].CellId != "testId" {
		fatalFail(errors.New("cell was not inserted with CreateCell"))
	}
}

func TestDeleteCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := generated.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIfNotNull(err, "could not create cell")
	if (*cm.Cells)[0].CellId != "testId" {
		fatalFail(errors.New("cell was not inserted with CreateCell"))
	}

	_, err = cm.DeleteCell(context.Background(), &request)
	failIfNotNull(err, "could not delete cell")
	if len(*cm.Cells) != 0 {
		fatalFail(errors.New("cell was not deleted with DeleteCell"))
	}
}

func TestListCells(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	cm.AppendCell(cell.Cell{CellId: "testId2"})
	cellList, err := cm.ListCells(context.Background(), &generated.ListCellsRequest{})
	failIfNotNull(err, "could not create cell")
	testId1Exists := false
	testId2Exists := false
	for _, cellId := range cellList.CellId {
		if cellId == "testId1" {
			testId1Exists = true
		}
		if cellId == "testId2" {
			testId2Exists = true
		}
	}
	if !testId1Exists || !testId2Exists {
		fatalFail(errors.New("cells were not returned from ListCells"))
	}
}
