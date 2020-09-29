package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"testing"
)

func TestSendUpdate(t *testing.T) {
	cm := objects.NewCellMaster()
	objIds := []string{"key"}
	newValues := []string{"value"}
	obj := generated.SingleObject{ObjectId: "testId", UpdateKey: objIds, NewValue: newValues}
	_, err := cm.SendUpdate(context.Background(), &obj)

	failIfNotNull(err, "could not update cellmaster")
	if (*cm.ObjectsToUpdate)[0].UpdateKey[0] == "key" {
		return
	}
	fatalFail(errors.New("object to update was not added to list"))
}

func TestGetCellState(t *testing.T) {
	cm := objects.NewCellMaster()

}

func TestIsAlive(t *testing.T) {
	cm := objects.NewCellMaster()

	request := generated.EmptyRequest{}
	_, err := cm.IsAlive(context.Background(), &request)

	failIfNotNull(err, "isAlive failed")
}