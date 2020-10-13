package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func createSingleObject(propertyKey string, newValue string, id string, cellId string) generated.SingleObject {
	objIds := []string{propertyKey}
	newValues := []string{newValue}
	return generated.SingleObject{ObjectId: id, UpdateKey: objIds, NewValue: newValues, CellId: cellId, PosY: 0, PosX: 0}
}

func TestSendUpdate(t *testing.T) {
	cm := objects.NewPlayer()
	obj := createSingleObject("key", "value", "key2", "cellId", )
	_, err := cm.RequestObjectMutation(context.Background(), &obj)

	failIfNotNull(err, "could not update cellmaster")
	if (*cm.MutatingObjects)[0].UpdateKey[0] == "key" {
		return
	}
	fatalFail(errors.New("object to update was not added to list"))
}

func TestReceiveCellMasterShip(t *testing.T) {
	cm := objects.NewPlayer()
	cl := generated.CellList{Cells: []*generated.Cell{{CellId: "cellid", PosY: 1, PosX: 1, Width: 1, Height: 1}}}
	cm.ReceiveCellMastership(context.Background(), &cl)
	if _, ok := (*cm.Cells)["cellid"]; ok {
		return
	}
	log.Fatalf("receive cell mastership does not work. ")
}

func TestRequestMutatingObjects(t *testing.T) {
	cm := objects.NewPlayer()
	cellID1Object := createSingleObject("key", "value", "key2", "cellId1")
	cellID2Object := createSingleObject("key2", "value2", "key2", "cellId2")
	cellID1Object2 := createSingleObject("key1", "value1", "key3", "cellId1")
	cm.AppendMutatingObject(cellID1Object)
	cm.AppendMutatingObject(cellID1Object2)
	cm.AppendMutatingObject(cellID2Object)

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

type PlayerClientWrapper struct {
	generated.PlayerClient
	object *generated.SingleObject
}

var player = PlayerClientWrapper{}
var player2 = PlayerClientWrapper{}

func (p PlayerClientWrapper) ReceiveMutatedObjects(ctx context.Context, in *generated.MultipleObjects, opts ...grpc.CallOption) (*generated.EmptyReply, error) {
	player.object = in.Objects[0]
	return nil, nil
}

func TestBroadcastMutatedObjects(t *testing.T) {
	cm := objects.NewPlayer()
	cellId1 := "cellId1"
	cellId2 := "cellId2"

	(*cm.SubscribedPlayers)[cellId1] = make([]generated.PlayerClient, 0)
	(*cm.SubscribedPlayers)[cellId2] = make([]generated.PlayerClient, 0)
	(*cm.SubscribedPlayers)[cellId1] = append((*cm.SubscribedPlayers)["id"], player)
	(*cm.SubscribedPlayers)[cellId2] = append((*cm.SubscribedPlayers)["id"], player2)

	objects := make([]*generated.SingleObject, 0)
	obj := createSingleObject("propertyKey", "newValeue", "objId", cellId1)
	objects = append(objects, &obj)

	multObjects := generated.MultipleObjects{Objects: objects}

	cm.BroadcastMutatedObjects(context.Background(), &multObjects)

	if player2.object != nil {
		fatalFail(errors.New("brodcast was sent to wrong cellID"))

	}

	if player.object.CellId == cellId1 {
		return
	}
	fatalFail(errors.New("broadcast updated objects failed"))

	//cm.SubscribedPlayers = append()
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
	cm := objects.NewPlayer()

	request := generated.EmptyRequest{}
	_, err := cm.IsAlive(context.Background(), &request)

	failIfNotNull(err, "isAlive failed")
}

func TestRequestObjectMutationSetToCorrectCellId(t *testing.T) {
	cm := objects.NewPlayer()
	cell := objects.NewCell("testCell")
	cell.PosX = 0
	cell.PosY = 0
	cell.Height = 100
	cell.Width = 100
	*cm.Cells = append(*cm.Cells, cell)

	_, err := cm.RequestObjectMutation(context.Background(), &generated.SingleObject{ObjectId: "test", PosX: 50, PosY: 50})
	failIfNotNull(err, "Failed RequestObjectMutation")

	if (*cm.MutatingObjects)[0].CellId != "testCell" {
		fatalFail(errors.New("mutating object set to wrong id"))
	}
}

//TODO figure out what to do with mutations outside of cellmaster cell
/*func TestRequestObjectMutationOutsideOfCell(t *testing.T) {
	cm := objects.NewPlayer()
	cell := objects.NewCell("testCell")
	cell.PosX = 0
	cell.PosY = 0
	cell.Height = 100
	cell.Width = 100
	*cm.Cells = append(*cm.Cells, cell)

	_, err := cm.RequestObjectMutation(context.Background(), &generated.SingleObject{ObjectId: "test", PosX: 150, PosY: 50})
	failIfNotNull(err, "Failed RequestObjectMutation")

	//TODO figure out what it should return
}*/
