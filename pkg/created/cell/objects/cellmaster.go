package objects

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
)

type CellMaster struct {
	objects.CellMasterServer
	ObjectsToUpdate *[]objects.SingleObject
}


func NewCellMaster() CellMaster {
	emptyList := make([]objects.SingleObject, 0)
	return CellMaster{ObjectsToUpdate: &emptyList}
}

func (cm *CellMaster) AppendObjectToUpdate(object objects.SingleObject) {
	*cm.ObjectsToUpdate = append(*cm.ObjectsToUpdate, object)
}

func (cm *CellMaster) SendUpdate(ctx context.Context, in *objects.SingleObject, ) (*objects.EmptyReply, error) {
	cm.AppendObjectToUpdate(*in)
	return &objects.EmptyReply{}, nil
}



func (cm *CellMaster) RequestMutatingObjects(ctx context.Context, in *objects.EmptyRequest) (*objects.MultipleObjects, error) {
	//cm.AppendObjectToUpdate(*in)
	return &objects.MultipleObjects{}, nil
}


//rpc RequestMutatingObjects (MultipleObjects) returns (MultipleObjects) {}



func (cm *CellMaster) IsAlive(ctx context.Context, in *objects.EmptyRequest) (*objects.EmptyReply, error) {
	return &objects.EmptyReply{}, nil
}