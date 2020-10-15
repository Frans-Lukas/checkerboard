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

	if player1.CellMaster.Port == -1 {
		fatalFail(errors.New("player was not updated with cellmaster"))
	}

	if player1.CellMaster.Port != 1337 || player1.CellMaster.Ip != "localhost" {
		fatalFail(errors.New("players cellmaster was not updated with correct variables"))
	}
}

func TestSendUpdateSingleObject(t *testing.T) {
	player1 := objects.NewPlayer()

	object := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"objKey1", "objKey2"}, NewValue:[]string{"value1", "value2"}}
	objectList := []*generated.SingleObject{&object}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.ReceiveMutatedObjects(context.Background(), &request)
	failIfNotNull(err, "could not fulfill sendUpdate")

	if len(*player1.MutatedObjects) == 0 {
		fatalFail(errors.New("objects not added to ObjectsToUpdate"))
	}

	if (*player1.MutatedObjects)[0].ObjectId != "object1" {
		fatalFail(errors.New("object does not have correct id"))
	}

	if !(((*player1.MutatedObjects)[0].UpdateKey[0] == "objKey1" && (*player1.MutatedObjects)[0].NewValue[0] == "value1") || ((*player1.MutatedObjects)[0].UpdateKey[1] == "objKey1" && (*player1.MutatedObjects)[0].NewValue[1] == "value1")) {
		fatalFail(errors.New("objects updated with incorrect values"))
	}

	if !(((*player1.MutatedObjects)[0].UpdateKey[0] == "objKey2" && (*player1.MutatedObjects)[0].NewValue[0] == "value2") || ((*player1.MutatedObjects)[0].UpdateKey[1] == "objKey2" && (*player1.MutatedObjects)[0].NewValue[1] == "value2")) {
		fatalFail(errors.New("objects updated with incorrect values"))
	}
}

func TestSendUpdateSingleObjectWrong(t *testing.T) {
	player1 := objects.NewPlayer()

	object := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"objKey1"}, NewValue:[]string{"value1", "value2"}}
	objectList := []*generated.SingleObject{&object}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.ReceiveMutatedObjects(context.Background(), &request)

	if err == nil {
		fatalFail(errors.New("objects are added when they should not be"))
	}
}

func TestSendUpdateMultipleObjects(t *testing.T) {
	player1 := objects.NewPlayer()

	object1 := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"1objKey1", "1objKey2"}, NewValue:[]string{"1value1", "1value2"}}
	object2 := generated.SingleObject{ObjectId: "object2", UpdateKey:[]string{"2objKey1", "2objKey2"}, NewValue:[]string{"2value1", "2value2"}}

	objectList := []*generated.SingleObject{&object1, &object2}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.ReceiveMutatedObjects(context.Background(), &request)
	failIfNotNull(err, "could not fulfill sendUpdate")

	if len(*player1.MutatedObjects) == 0 {
		fatalFail(errors.New("objects not added to ObjectsToUpdate"))
	}

	object1Updated := generated.SingleObject{}
	object2Updated := generated.SingleObject{}

	if (*player1.MutatedObjects)[0].ObjectId == "object1" {
		object1Updated = (*player1.MutatedObjects)[0]
	} else if (*player1.MutatedObjects)[0].ObjectId == "object2" {
		object2Updated = (*player1.MutatedObjects)[0]
	} else {
		fatalFail(errors.New("object with invalid objectId"))
	}

	if (*player1.MutatedObjects)[1].ObjectId == "object1" && object1Updated.ObjectId != "object1" {
		object1Updated = (*player1.MutatedObjects)[1]
	} else if (*player1.MutatedObjects)[1].ObjectId == "object2" && object1Updated.ObjectId != "object2" {
		object2Updated = (*player1.MutatedObjects)[1]
	} else {
		fatalFail(errors.New("object with invalid objectId"))
	}

	if !((object1Updated.UpdateKey[0] == "1objKey1" && object1Updated.NewValue[0] == "1value1") || (object1Updated.UpdateKey[1] == "1objKey1" && object1Updated.NewValue[1] == "1value1")) {
		fatalFail(errors.New("objects updated with incorrect values"))
	}

	if !((object1Updated.UpdateKey[0] == "1objKey2" && object1Updated.NewValue[0] == "1value2") || (object1Updated.UpdateKey[1] == "1objKey2" && object1Updated.NewValue[1] == "1value2")) {
		fatalFail(errors.New("objects updated with incorrect values"))
	}

	/*if player1.MutatedObjects["object1"]["1objKey1"] != "1value1" || player1.MutatedObjects["object1"]["1objKey2"] != "1value2" {
		fatalFail(errors.New("objects updated with incorrect values"))
	}*/

	if !((object2Updated.UpdateKey[0] == "2objKey1" && object2Updated.NewValue[0] == "2value1") || (object2Updated.UpdateKey[1] == "2objKey1" && object2Updated.NewValue[1] == "2value1")) {
		fatalFail(errors.New("objects updated with incorrect values"))
	}

	if !((object2Updated.UpdateKey[0] == "2objKey2" && object2Updated.NewValue[0] == "2value2") || (object2Updated.UpdateKey[1] == "2objKey2" && object2Updated.NewValue[1] == "2value2")) {
		fatalFail(errors.New("objects updated with incorrect values"))
	}

	/*if player1.MutatedObjects["object2"]["2objKey1"] != "2value1" || player1.MutatedObjects["object2"]["2objKey2"] != "2value2" {
		fatalFail(errors.New("objects updated with incorrect values"))
	}*/
}

func TestSendUpdateMultipleObjectsWrong(t *testing.T) {
	player1 := objects.NewPlayer()

	object1 := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"1objKey1", "1objKey2"}, NewValue:[]string{"1value1", "1value2"}}
	object2 := generated.SingleObject{ObjectId: "object2", UpdateKey:[]string{"2objKey1"}, NewValue:[]string{"2value1", "2value2"}}

	objectList := []*generated.SingleObject{&object1, &object2}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.ReceiveMutatedObjects(context.Background(), &request)
	if err == nil || len(*player1.MutatedObjects) != 0{
		fatalFail(errors.New("objects are added when they should not be"))
	}
}

//TODO figure out how to change this test
/*func TestSendUpdateMultipleTimes(t *testing.T) {
	player1 := objects.NewPlayer()

	object1 := generated.SingleObject{ObjectId: "object1", UpdateKey:[]string{"1objKey1", "1objKey2"}, NewValue:[]string{"1value1", "1value2"}}
	object2 := generated.SingleObject{ObjectId: "object2", UpdateKey:[]string{"2objKey1", "2objKey2"}, NewValue:[]string{"2value1", "2value2"}}

	objectList := []*generated.SingleObject{&object1, &object2}

	request := generated.MultipleObjects{Objects: objectList}

	_, err := player1.ReceiveMutatedObjects(context.Background(), &request)
	failIfNotNull(err, "could not fulfill sendUpdate")

	object2 = generated.SingleObject{ObjectId: "object2", UpdateKey:[]string{"2objKey2", "2objKey3"}, NewValue:[]string{"2newValue2", "2value3"}}
	object3 := generated.SingleObject{ObjectId: "object3", UpdateKey:[]string{"3objKey1", "3objKey2"}, NewValue:[]string{"3value1", "3value2"}}

	objectList = []*generated.SingleObject{&object2, &object3}

	request = generated.MultipleObjects{Objects: objectList}

	_, err = player1.ReceiveMutatedObjects(context.Background(), &request)
	failIfNotNull(err, "could not fulfill sendUpdate")

	if len(*player1.MutatedObjects) != 3 {
		fatalFail(errors.New("objects not added to ObjectsToUpdate"))
	}

	object2Updated := generated.SingleObject{}

	if (*player1.MutatedObjects)[0].ObjectId == "object2" {
		object2Updated = (*player1.MutatedObjects)[0]
	} else if (*player1.MutatedObjects)[1].ObjectId == "object2" {
		object2Updated = (*player1.MutatedObjects)[1]
	} else if (*player1.MutatedObjects)[2].ObjectId == "object2" {
		object2Updated = (*player1.MutatedObjects)[2]
	}


	if (object2Updated.UpdateKey[0] == "2objKey1" && object2Updated.NewValue[0] == "2value1") || (object2Updated.UpdateKey[1] == "2objKey1" && object2Updated.NewValue[1] == "2value1") {
		fatalFail(errors.New("objects updated with incorrect values"))
	}

	if (object2Updated.UpdateKey[0] == "2objKey2" && object2Updated.NewValue[0] == "2value2") || (object2Updated.UpdateKey[1] == "2objKey2" && object2Updated.NewValue[1] == "2value2") {
		fatalFail(errors.New("objects updated with incorrect values"))
	}
}*/


