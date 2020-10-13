package objects

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
)

type Client struct {
	//TODO decide what to do with player and client
	Port int32
	Ip   string
	// 0 = lowest trust level UINT32_MAX = highest trust level
	TrustLevel uint32
}

type Player struct {
	generated.PlayerServer
	CellMaster        Client
	MutatedObjects    map[string]map[string]string
	MutatingObjects   *[]generated.SingleObject
	SubscribedPlayers *map[string][]generated.PlayerClient
	Cells             *[]Cell
}

func NewPlayer() Player {
	emptyObjectList := make([]generated.SingleObject, 0)
	cells := make([]Cell, 0)
	emptyPlayerMap := make(map[string][]generated.PlayerClient, 0)
	return Player{CellMaster: Client{Port: -1, Ip: "none"}, MutatedObjects: map[string]map[string]string{}, SubscribedPlayers: &emptyPlayerMap, MutatingObjects: &emptyObjectList, Cells: &cells}
}

func (player *Player) UpdateCellMaster(
	ctx context.Context, in *generated.NewCellMaster,
) (*generated.EmptyReply, error) {
	player.CellMaster = Client{Ip: in.Ip, Port: in.Port}
	return &generated.EmptyReply{}, nil
}

func (player *Player) ReceiveMutatedObjects(
	ctx context.Context, in *generated.MultipleObjects,
) (*generated.EmptyReply, error) {
	for _, object := range in.Objects {
		if len(object.NewValue) != len(object.UpdateKey) {
			return &generated.EmptyReply{}, errors.New("not as many values as keys")
		}
	}

	for _, object := range in.Objects {

		if player.MutatedObjects[object.ObjectId] == nil {
			player.MutatedObjects[object.ObjectId] = map[string]string{}
		}

		objectsToUpdate := player.MutatedObjects[object.ObjectId]
		for index, key := range object.UpdateKey {
			objectsToUpdate[key] = object.NewValue[index]
		}
	}

	return &generated.EmptyReply{}, nil
}

func (cm *Player) AppendMutatingObject(object generated.SingleObject) {
	*cm.MutatingObjects = append(*cm.MutatingObjects, object)
}

func (cm *Player) RequestObjectMutation(ctx context.Context, in *generated.SingleObject) (*generated.EmptyReply, error) {
	println("Received object mutation request for type: ", in.ObjectType)
	// TODO: make sure overlapping objects
	for _, cell := range *cm.Cells {
		println("iterating cell with id ", cell.CellId)
		if cell.CollidesWith(&cellmanager.Position{PosY: in.PosY, PosX: in.PosX}) {
			println("object collides with cell with id ", cell.CellId)
			in.CellId = cell.CellId
			break
		}
	}
	cm.AppendMutatingObject(*in)
	return &generated.EmptyReply{}, nil
}

func (cm *Player) RequestMutatingObjects(ctx context.Context, in *generated.Cell) (*generated.MultipleObjects, error) {
	mutatingObjects := make([]*generated.SingleObject, 0)

	for index, object := range *cm.MutatingObjects {
		if object.CellId == in.CellId {
			mutatingObjects = append(mutatingObjects, &(*cm.MutatingObjects)[index])
		}
	}

	return &generated.MultipleObjects{Objects: mutatingObjects}, nil
}

func (cm *Player) BroadcastMutatedObjects(ctx context.Context, in *generated.MultipleObjects) (*generated.EmptyReply, error) {
	for objectIndex, object := range (*in).Objects {
		println("checking cell with id ", object.CellId)
		if playerList, ok := (*cm.SubscribedPlayers)[object.CellId]; ok {
			println("broadcasting to cell with id ", object.CellId)
			for _, player := range playerList {
				println("sending updated objects to player: ", player)
				err := cm.SendObjectUpdateToPlayer(player, ctx, (*in).Objects[objectIndex])
				if err != nil {
					return nil, err
				}

			}
		}
	}
	return &generated.EmptyReply{}, nil
}

func (cm *Player) SendObjectUpdateToPlayer(player generated.PlayerClient, ctx context.Context, object *generated.SingleObject) (error) {
	_, err := player.ReceiveMutatedObjects(ctx, &generated.MultipleObjects{Objects: []*generated.SingleObject{object}})
	return err
}

//TODO implement cell state
/*func (cm *CellMaster) GetCellState(ctx context.Context, in *objects.Cell) (*objects.MultipleObjects, error) {
	return &objects.MultipleObjects{}, nil
}*/

func (cm *Player) IsAlive(ctx context.Context, in *generated.EmptyRequest) (*generated.EmptyReply, error) {
	return &generated.EmptyReply{}, nil
}
