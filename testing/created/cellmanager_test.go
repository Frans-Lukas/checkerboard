package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cellmanager"
	"github.com/Frans-Lukas/checkerboard/pkg/generated"
	"testing"
)

func TestCreateCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := generated.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIf(err, "could not create cell")
	if (*cm.Cells)[0].CellId != "testId" {
		fatalFail(errors.New("cell was not inserted"))
	}
}
