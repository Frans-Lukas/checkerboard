package objects

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

type Client struct {
	//TODO decide what to do with player and client
	Port int32
	Ip   string
	// 0 = lowest trust level UINT32_MAX = highest trust level
	TrustLevel uint32
}

type CellMasterConnection struct {
	CellMaster *generated.PlayerClient
	Connection *grpc.ClientConn
}

type Player struct {
	generated.PlayerServer
	CellMasterConnection
	MutatedObjects  *[]generated.SingleObject
	MutatingObjects *[]generated.SingleObject

	//map of cellid map of playerid
	SubscribedPlayers    *map[string]map[string]*generated.PlayerClient
	Cells                *map[string]Cell
	splitCellRequirement int
	splitCheckInterval   int
}

func NewPlayer(splitCellRequirement int, splitCheckInterval int) Player {
	emptyObjectList := make([]generated.SingleObject, 0)
	cells := make(map[string]Cell, 0)
	emptyPlayerMap := make(map[string]map[string]*generated.PlayerClient, 0)
	mutatedObjects := make([]generated.SingleObject, 0)
	return Player{
		MutatedObjects:       &mutatedObjects,
		SubscribedPlayers:    &emptyPlayerMap,
		MutatingObjects:      &emptyObjectList,
		Cells:                &cells,
		splitCellRequirement: splitCellRequirement,
		splitCheckInterval:   splitCheckInterval,
	}
}

func (player *Player) ReceiveMutatedObjects(
	ctx context.Context, in *generated.MultipleObjects,
) (*generated.EmptyReply, error) {
	if constants.DebugMode {
		println("Received mutated object")
	}
	for _, object := range in.Objects {
		if len(object.NewValue) != len(object.UpdateKey) {
			return &generated.EmptyReply{}, errors.New("not as many values as keys")
		}
	}

	for _, object := range in.Objects {
		*player.MutatedObjects = append(*player.MutatedObjects, *object)
	}

	return &generated.EmptyReply{}, nil
}

func (cm *Player) AppendMutatingObject(object generated.SingleObject) {
	if constants.DebugMode {
		println("Appending object with cellid ", object.CellId)
	}
	*cm.MutatingObjects = append(*cm.MutatingObjects, object)
}

func (cm *Player) ReceiveCellMastership(ctx context.Context, in *generated.CellList) (*generated.EmptyReply, error) {
	if constants.DebugMode {
		println("Received cell mastership")
	}
	for _, cell := range in.Cells {
		if ownedCell, ok := (*cm.Cells)[cell.CellId]; ok {
			// cm is already aware of mastership over cell, update cell status
			ownedCell.PosX = cell.PosX
			ownedCell.PosY = cell.PosY
			ownedCell.Height = cell.Height
			ownedCell.Width = cell.Width

			(*cm.Cells)[cell.CellId] = ownedCell
		} else {
			(*cm.Cells)[cell.CellId] = Cell{CellId: cell.CellId, PosX: cell.PosX, PosY: cell.PosY, Width: cell.Width, Height: cell.Height}
		}
	}
	return &generated.EmptyReply{}, nil
}

func (cm *Player) RequestObjectMutation(ctx context.Context, in *generated.SingleObject) (*generated.EmptyReply, error) {
	if constants.DebugMode {
		println("Received object mutation request for type: ", in.ObjectType)
	}
	// TODO: make sure overlapping objects
	for _, cell := range *cm.Cells {
		if constants.DebugMode {
			println("iterating cell with id ", cell.CellId)
		}
		if cell.CollidesWith(&cellmanager.Position{PosY: in.PosY, PosX: in.PosX}) {
			if constants.DebugMode {
				println("object collides with cell with id ", cell.CellId)
			}
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
		if constants.DebugMode {
			println("checking cell with id ", object.CellId)
		}
		if playerList, ok := (*cm.SubscribedPlayers)[object.CellId]; ok {
			if constants.DebugMode {
				println("checking playerlist of size ", len(playerList))
				println("broadcasting to cell with id ", object.CellId)
			}
			for _, player := range playerList {
				if constants.DebugMode {
					println("sending updated objects to player: ", player)
				}
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
	if constants.DebugMode {
		println("Sending object update to player ")
	}
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

func (cm *Player) ChangedCellMaster(ctx context.Context, in *generated.ChangedCellMasterRequest) (*generated.ChangedCellMasterReply, error) {
	cm.CellMaster = nil
	cm.Connection.Close()
	return &generated.ChangedCellMasterReply{}, nil
}

func (cm *Player) SubscribePlayer(ctx context.Context, in *generated.PlayerInfo) (*generated.SubscriptionReply, error) {
	subscribedToCell := false
	if constants.DebugMode {
		println("Subscribing player: ", in.Port)
	}
	for _, cell := range *cm.Cells {
		if cell.CollidesWith(&cellmanager.Position{PosX: in.PosX, PosY: in.PosY}) {
			if _, exists := (*cm.SubscribedPlayers)[cell.CellId]; !exists {
				(*cm.SubscribedPlayers)[cell.CellId] = map[string]*generated.PlayerClient{}
			}

			subscribers := (*cm.SubscribedPlayers)[cell.CellId]

			if _, exists := (subscribers)[in.Ip+":"+strconv.Itoa(int(in.Port))]; !exists {

				conn, err2 := grpc.Dial(ToAddress(in.Ip, in.Port), grpc.WithInsecure(), grpc.WithBlock())
				if err2 != nil {
					if constants.DebugMode {
						println("did not connect to subscriber: %v", err2)
					}
					return &generated.SubscriptionReply{Succeeded: false}, errors.New("could not connect")
				}
				if constants.DebugMode {
					println("Actually subscribing player: ", in.Port)
				}

				subscriberConn := generated.NewPlayerClient(conn)
				subscribers[in.Ip+":"+strconv.Itoa(int(in.Port))] = &subscriberConn
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

func (cm *Player) ShouldSplitCell() (shouldSplit bool, cellId string) {
	for cellId, playerList := range *cm.SubscribedPlayers {
		// only split one cell at a time
		return len(playerList) > cm.splitCellRequirement, cellId
	}

	return false, ""
}

func (cm *Player) SplitCellLoop(client *cellmanager.CellManagerClient) {
	for {
		shouldSplit, cellId := cm.ShouldSplitCell()

		if shouldSplit {
			cm.SplitCell(client, cellId)
		}

		time.Sleep(time.Second * time.Duration(cm.splitCheckInterval))
	}
}

func (cm *Player) SplitCell(client *cellmanager.CellManagerClient, cellID string) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := (*client).DivideCell(ctx, &cellmanager.CellRequest{CellId: cellID})
	if err != nil {
		println("Failed to split cell")
		return
	}

	cm.DesubscribePlayers(ctx)
	// reset subscribed players map
	cm.SubscribedPlayers = new(map[string]map[string]*generated.PlayerClient)

}

func (cm *Player) DesubscribePlayers(ctx context.Context) {
	for _, playerMap := range *cm.SubscribedPlayers {
		for _, player := range playerMap {
			(*player).ChangedCellMaster(ctx, &generated.ChangedCellMasterRequest{})
		}
	}
}
