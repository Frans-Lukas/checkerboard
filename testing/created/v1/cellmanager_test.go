package v1

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/v1/cellmanager"
	pb "github.com/Frans-Lukas/checkerboard/pkg/generated/v1"
	"testing"
)

func TestCreateCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := pb.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIf(err, "could not create cell")
	if (*cm.Cells)[0].CellId != "testId" {
		fatalFail(errors.New("cell was not inserted"))
	}
}
