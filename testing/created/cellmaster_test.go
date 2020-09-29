package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"testing"
)

func createSingleObject(objId string, newValue string, id string) generated.SingleObject {
	objIds := []string{objId}
	newValues := []string{newValue}
	return generated.SingleObject{ObjectId: id, UpdateKey: objIds, NewValue: newValues}
}

func TestSendUpdate(t *testing.T) {
	cm := objects.NewCellMaster()
	obj := createSingleObject("key", "value", "key2")
	_, err := cm.SendUpdate(context.Background(), &obj)

	failIfNotNull(err, "could not update cellmaster")
	if (*cm.ObjectsToUpdate)[0].UpdateKey[0] == "key" {
		return
	}
	fatalFail(errors.New("object to update was not added to list"))
}

func TestRequestMutatingObjects(t *testing.T) {
	cm := objects.NewCellMaster()
	emptyRequest := generated.EmptyRequest{}
	obj := createSingleObject("key", "value", "key2")
	cm.AppendObjectToUpdate(obj)

	mutatingObjects, err := cm.RequestMutatingObjects(context.Background(), &emptyRequest)

	//if (*mutatingObjects)[0]

	failIfNotNull(err, "could not update cellmaster")
	if (*cm.ObjectsToUpdate)[0].UpdateKey[0] == "key" {
		return
	}
	fatalFail(errors.New("object to update was not added to list"))
}
