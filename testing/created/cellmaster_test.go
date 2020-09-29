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

//TODO put back when cell state is implemented
/*func TestGetCellState(t *testing.T) {
	cm := objects.NewCellMaster()

	object1 := generated.SingleObject{CellId: "testCell", ObjectId: "object1", UpdateKey:[]string{"1objKey1", "1objKey2"}, NewValue:[]string{"1value1", "1value2"}}
	object2 := generated.SingleObject{CellId: "testCell", ObjectId: "object2", UpdateKey:[]string{"2objKey1", "2objKey2"}, NewValue:[]string{"2value1", "2value2"}}
	object3 := generated.SingleObject{CellId: "testCell2", ObjectId: "object3", UpdateKey:[]string{"3objKey1", "3objKey2"}, NewValue:[]string{"3value1", "3value2"}}

	_, err := cm.SendUpdate(context.Background(), &object1)
	failIfNotNull(err, "could not sendUpdate")
	_, err = cm.SendUpdate(context.Background(), &object2)
	failIfNotNull(err, "could not sendUpdate")
	_, err = cm.SendUpdate(context.Background(), &object3)
	failIfNotNull(err, "could not sendUpdate")

	cellRequest := generated.Cell{CellId: "testCell"}

	cellState, err := cm.GetCellState(context.Background(), &cellRequest)
	failIfNotNull(err, "could not getCellState")

	if len(cellState.Objects) != 2 {
		fatalFail(errors.New("got incorrect number of objects back"))
	}

	if !(cellState.Objects[0].ObjectId == "object1" || cellState.Objects[1].ObjectId == "object1") {
		fatalFail(errors.New("object1 not added to cell state"))
	}

	if !(cellState.Objects[0].ObjectId == "object2" || cellState.Objects[1].ObjectId == "object2") {
		fatalFail(errors.New("object2 not added to cell state"))
	}
}*/

func TestIsAlive(t *testing.T) {
	cm := objects.NewCellMaster()

	request := generated.EmptyRequest{}
	_, err := cm.IsAlive(context.Background(), &request)

	failIfNotNull(err, "isAlive failed")
}