package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"testing"
)

func createSingleObject(objId string, newValue string, id string, cellId string) generated.SingleObject {
	objIds := []string{objId}
	newValues := []string{newValue}
	return generated.SingleObject{ObjectId: id, UpdateKey: objIds, NewValue: newValues, CellId: cellId}
}

func TestSendUpdate(t *testing.T) {
	cm := objects.NewCellMaster()
	obj := createSingleObject("key", "value", "key2", "cellId")
	_, err := cm.SendUpdate(context.Background(), &obj)

	failIfNotNull(err, "could not update cellmaster")
	if (*cm.ObjectsToUpdate)[0].UpdateKey[0] == "key" {
		return
	}
	fatalFail(errors.New("object to update was not added to list"))
}

func TestRequestMutatingObjects(t *testing.T) {
	cm := objects.NewCellMaster()
	cellID1Object := createSingleObject("key", "value", "key2", "cellId1")
	cellID2Object := createSingleObject("key2", "value2", "key2", "cellId2")
	cellID1Object2 := createSingleObject("key1", "value1", "key3", "cellId1")
	cm.AppendObjectToUpdate(cellID1Object)
	cm.AppendObjectToUpdate(cellID1Object2)
	cm.AppendObjectToUpdate(cellID2Object)

	cellId := generated.Cell{CellId: "cellId1"}
	mutatingObjects, err := cm.RequestMutatingObjects(context.Background(), &cellId)
	failIfNotNull(err, "could not update cellmaster")
	containsCorrectObject := 0
	for _, value := range (*mutatingObjects).Objects {
		if value.CellId == "cellId1" {
			containsCorrectObject++
		} else {
			fatalFail(errors.New("RequestMutatingObjects returned invalid cellId object"))
		}
	}

	if containsCorrectObject != 2 {
		fatalFail(errors.New("RequestMutatingObjects did not return correct object"))
	}
}

func TestBroadcastMutatedObjects(t *testing.T) {
	//cm := objects.NewCellMaster()
	//cm.SubscribedPlayers = append()
}

func TestGetCellState(t *testing.T) {
	//cm := objects.NewCellMaster()

}

func TestIsAlive(t *testing.T) {
	cm := objects.NewCellMaster()

	request := generated.EmptyRequest{}
	_, err := cm.IsAlive(context.Background(), &request)

	failIfNotNull(err, "isAlive failed")
}
