package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"testing"
)

func TestUpdateCellMaster(t *testing.T) {
	player1 := objects.NewPlayer()
	request := generated.NewCellMaster{Ip: "localhost", Port: 1337}
	_, err := player1.UpdateCellMaster(context.Background(), &request)
	failIfNotNull(err, "could not update cellmaster")

	if player1.Cellmaster.Port == -1 {
		fatalFail(errors.New("player was not updated with cellmaster"))
	}

	if player1.Cellmaster.Port != 1337 || player1.Cellmaster.Ip != "localhost" {
		fatalFail(errors.New("players cellmaster was not updated with correct variables"))
	}
}

func TestSendSingleObjectUpdate(t *testing.T) {
	player1 := objects.NewPlayer()

	object := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"objKey1", "objKey2"}, NewValue:[]string{"value1", "value2"}}
	objectList := []*generated.SingleObject{&object}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.SendUpdate(context.Background(), &request)
	failIfNotNull(err, "could not sendUpdate")

	if len(player1.ObjectsToUpdate) == 0 {
		fatalFail(errors.New("objects not added to ObjectsToUpdate"))
	}

	if player1.ObjectsToUpdate["object1"]["objKey1"] != "value1" || player1.ObjectsToUpdate["object1"]["objKey2"] != "value2" {
		fatalFail(errors.New("objects updated with incorrect values"))
	}
}

func TestSendSingleObjectWrongUpdate(t *testing.T) {
	player1 := objects.NewPlayer()

	object := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"objKey1"}, NewValue:[]string{"value1", "value2"}}
	objectList := []*generated.SingleObject{&object}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.SendUpdate(context.Background(), &request)

	if err == nil {
		fatalFail(errors.New("objects are added when they should not be"))
	}
}