package created

import (
	"context"
	"errors"
	"fmt"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"log"
	"net"
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

//TODO put back when cell state is implemented
/*func TestGetCellState(t *testing.T) {
	cm := objects.PlayerInfo()

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
	(*cm.Cells)["testcell"] = cell

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

func TestSubscribePlayerToSingleCell(t *testing.T) {
	cm := objects.NewPlayer()
	cell1 := objects.NewCell("cell1")
	cell1.PosX = 0
	cell1.PosY = 0
	cell1.Height = 100
	cell1.Width = 100
	(*cm.Cells)[cell1.CellId] = cell1
	cell2 := objects.NewCell("cell2")
	cell2.PosX = 50
	cell2.PosY = 0
	cell2.Height = 100
	cell2.Width = 100
	(*cm.Cells)[cell2.CellId] = cell2

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	playerServer := grpc.NewServer()
	subscriber := objects.NewPlayer()
	generated.RegisterPlayerServer(playerServer, &subscriber)
	go func() {
		if err := playerServer.Serve(lis); err != nil && err.Error() != "the server has been stopped" {
			log.Fatalf("failed to serve %v", err)
		}
	}()

	res, err := cm.SubscribePlayer(context.Background(), &generated.PlayerInfo{Ip:"localhost", Port: 8888, PosX:40, PosY:40})
	failIfNotNull(err, "Received error on subscription: ")

	if !res.Succeeded {
		fatalFail(errors.New("succeeded == false but no error"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell1.CellId]["localhost:8888"] ; !subscribed {
		fatalFail(errors.New("not subscribed to cell1"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell2.CellId]["localhost:8888"] ; subscribed {
		fatalFail(errors.New("subscribed to cell2"))
	}

	playerServer.GracefulStop()
}

func TestSubscribePlayerToMultipleCells(t *testing.T) {
	cm := objects.NewPlayer()
	cell1 := objects.NewCell("cell1")
	cell1.PosX = 0
	cell1.PosY = 0
	cell1.Height = 100
	cell1.Width = 100
	(*cm.Cells)[cell1.CellId] = cell1
	cell2 := objects.NewCell("cell2")
	cell2.PosX = 50
	cell2.PosY = 0
	cell2.Height = 100
	cell2.Width = 100
	(*cm.Cells)[cell2.CellId] = cell2

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	playerServer := grpc.NewServer()
	subscriber := objects.NewPlayer()
	generated.RegisterPlayerServer(playerServer, &subscriber)
	go func() {
		if err := playerServer.Serve(lis); err != nil && err.Error() != "the server has been stopped" {
			log.Fatalf("failed to serve %v", err)
		}
	}()

	res, err := cm.SubscribePlayer(context.Background(), &generated.PlayerInfo{Ip:"localhost", Port: 8888, PosX:60, PosY:60})

	if !res.Succeeded {
		fatalFail(errors.New("succeeded == false but no error"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell1.CellId]["localhost:8888"] ; !subscribed {
		fatalFail(errors.New("not subscribed to cell1"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell2.CellId]["localhost:8888"] ; !subscribed {
		fatalFail(errors.New("not subscribed to cell2"))
	}

	playerServer.GracefulStop()
}

func TestSubscribePlayerOutsideCells(t *testing.T) {
	cm := objects.NewPlayer()
	cell1 := objects.NewCell("cell1")
	cell1.PosX = 0
	cell1.PosY = 0
	cell1.Height = 100
	cell1.Width = 100
	(*cm.Cells)[cell1.CellId] = cell1
	cell2 := objects.NewCell("cell2")
	cell2.PosX = 50
	cell2.PosY = 0
	cell2.Height = 100
	cell2.Width = 100
	(*cm.Cells)[cell2.CellId] = cell2

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	playerServer := grpc.NewServer()
	subscriber := objects.NewPlayer()
	generated.RegisterPlayerServer(playerServer, &subscriber)
	go func() {
		if err := playerServer.Serve(lis); err != nil && err.Error() != "the server has been stopped" {
			log.Fatalf("failed to serve %v", err)
		}
	}()

	res, err := cm.SubscribePlayer(context.Background(), &generated.PlayerInfo{Ip:"localhost", Port: 8888, PosX:200, PosY:200})
	if err == nil {
		fatalFail(errors.New("did not receive error on failed subscription"))
	}

	if res.Succeeded {
		fatalFail(errors.New("succeeded == true but there is an error"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell1.CellId]["localhost:8888"] ; subscribed {
		fatalFail(errors.New("subscribed to cell1"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell2.CellId]["localhost:8888"] ; subscribed {
		fatalFail(errors.New("subscribed to cell2"))
	}

	playerServer.GracefulStop()
}

func TestSubscribeUnConnectablePlayerCells(t *testing.T) {
	cm := objects.NewPlayer()
	cell1 := objects.NewCell("cell1")
	cell1.PosX = 0
	cell1.PosY = 0
	cell1.Height = 100
	cell1.Width = 100
	(*cm.Cells)[cell1.CellId] = cell1
	cell2 := objects.NewCell("cell2")
	cell2.PosX = 50
	cell2.PosY = 0
	cell2.Height = 100
	cell2.Width = 100
	(*cm.Cells)[cell2.CellId] = cell2

	res, err := cm.SubscribePlayer(context.Background(), &generated.PlayerInfo{Ip:"localhost", Port: 8888, PosX:200, PosY:200})
	if err == nil {
		fatalFail(errors.New("did not receive error on failed subscription"))
	}

	if res.Succeeded {
		fatalFail(errors.New("succeeded == true but there is an error"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell1.CellId]["localhost:8888"] ; subscribed {
		fatalFail(errors.New("subscribed to cell1"))
	}

	if _, subscribed := (*cm.SubscribedPlayers)[cell2.CellId]["localhost:8888"] ; subscribed {
		fatalFail(errors.New("subscribed to cell2"))
	}
}