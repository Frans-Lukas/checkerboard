package objects

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"strconv"
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
	SubscribedPlayers *map[string]map[string]*generated.PlayerClient
	Cells             *map[string]Cell
}

func NewPlayer() Player {
	emptyObjectList := make([]generated.SingleObject, 0)
	cells := make(map[string]Cell, 0)
	emptyPlayerMap := make(map[string]map[string]*generated.PlayerClient, 0)
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
	println("Received mutated object")
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
	println("Appending object with cellid ", object.CellId)
	*cm.MutatingObjects = append(*cm.MutatingObjects, object)
}

func (cm *Player) ReceiveCellMastership(ctx context.Context, in *generated.CellList) (*generated.EmptyReply, error) {
	println("Received cell mastership")
	for _, cell := range in.Cells {
		if _, ok := (*cm.Cells)[cell.CellId]; ok {
			// cm is already aware of mastership over cell
		} else {
			(*cm.Cells)[cell.CellId] = Cell{CellId: cell.CellId, PosX: cell.PosX, PosY: cell.PosY, Width: cell.Width, Height: cell.Height}
		}
	}
	return &generated.EmptyReply{}, nil
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
				err := cm.SendObjectUpdateToPlayer(*player, ctx, (*in).Objects[objectIndex])
				if err != nil {
					return nil, err
				}

			}
		}
	}
	return &generated.EmptyReply{}, nil
}

func (cm *Player) SendObjectUpdateToPlayer(player generated.PlayerClient, ctx context.Context, object *generated.SingleObject) (error) {
	println("Sending object update to player ")
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

func (cm *Player) SubscribePlayer(ctx context.Context, in *generated.PlayerInfo) (*generated.SubscriptionReply, error) {
	subscribedToCell := false
	for _, cell := range *cm.Cells{
		if cell.CollidesWith(&cellmanager.Position{PosX: in.PosX, PosY: in.PosY}) {
			if _, exists := (*cm.SubscribedPlayers)[cell.CellId]; !exists {
				(*cm.SubscribedPlayers)[cell.CellId] = map[string]*generated.PlayerClient{}
			}

			subscribers := (*cm.SubscribedPlayers)[cell.CellId]

			if _, exists := (subscribers)[in.Ip + ":" + strconv.Itoa(int(in.Port))]; !exists {


				conn, err2 := grpc.Dial(ToAddress(in.Ip, in.Port), grpc.WithInsecure(), grpc.WithBlock())
				if err2 != nil {
					println("did not connect to subscriber: %v", err2)
					return &generated.SubscriptionReply{Succeeded: false}, errors.New("could not connect")
				}

				subscriberConn := generated.NewPlayerClient(conn)
				subscribers[in.Ip + ":" + strconv.Itoa(int(in.Port))] = &subscriberConn
				subscribedToCell = true
			}
		}
	}
	if !subscribedToCell {
		return &generated.SubscriptionReply{Succeeded: false}, errors.New("no colliding cell")
	} else {
		return &generated.SubscriptionReply{Succeeded: true}, nil
	}
}