package objects

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
)

type CellMaster struct {
	objects.CellMasterServer
	ObjectsToUpdate   *[]objects.SingleObject
	SubscribedPlayers *map[string][]objects.PlayerClient
}

func NewCellMaster() CellMaster {
	emptyObjectList := make([]objects.SingleObject, 0)
	emptyPlayerMap := make(map[string][]objects.PlayerClient, 0)
	return CellMaster{ObjectsToUpdate: &emptyObjectList, SubscribedPlayers: &emptyPlayerMap}
}

func (cm *CellMaster) AppendObjectToUpdate(object objects.SingleObject) {
	*cm.ObjectsToUpdate = append(*cm.ObjectsToUpdate, object)
}

func (cm *CellMaster) SendUpdate(ctx context.Context, in *objects.SingleObject, ) (*objects.EmptyReply, error) {
	cm.AppendObjectToUpdate(*in)
	return &objects.EmptyReply{}, nil
}

func (cm *CellMaster) RequestMutatingObjects(ctx context.Context, in *objects.Cell) (*objects.MultipleObjects, error) {
	mutatingObjects := make([]*objects.SingleObject, 0)

	for index, object := range *cm.ObjectsToUpdate {
		if object.CellId == in.CellId {
			mutatingObjects = append(mutatingObjects, &(*cm.ObjectsToUpdate)[index])
		}
	}

	return &objects.MultipleObjects{Objects: mutatingObjects}, nil
}

//TODO: Test this fucker
func (cm *CellMaster) BroadcastMutatedObjects(ctx context.Context, in *objects.MultipleObjects) (*objects.EmptyReply, error) {
	for objectIndex, object := range (*in).Objects {
		if playerList, ok := (*cm.SubscribedPlayers)[object.CellId]; ok {
			for _, player := range playerList {
				err := cm.SendObjectUpdateToPlayer(&player, ctx, (*in).Objects[objectIndex])
				if err != nil {
					return nil, err
				}

			}
		}
	}
	return &objects.EmptyReply{}, nil
}

func (cm *CellMaster) SendObjectUpdateToPlayer(player *objects.PlayerClient, ctx context.Context, object *objects.SingleObject) (error) {
	_, err := (*player).SendUpdate(ctx, &objects.MultipleObjects{Objects: []*objects.SingleObject{object}})
	return err
}

func (cm *CellMaster) IsAlive(ctx context.Context, in *objects.EmptyRequest) (*objects.EmptyReply, error) {
	return &objects.EmptyReply{}, nil
}
