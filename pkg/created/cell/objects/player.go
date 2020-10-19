package objects

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"strconv"
	"sync"
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

type PlayerInfoClient struct {
	generated.PlayerClient
	Port     int
	Ip       string
	ObjectId string
}

type Player struct {
	generated.PlayerServer
	*CellMasterConnection

	// must be set explicitly
	Port int
	Ip   string

	PosX     int64
	PosY     int64
	ObjectId string

	MutatedObjects  *[]generated.SingleObject
	MutatingObjects *[]generated.SingleObject

	//map of cellid map of playerid
	SubscribedPlayers *map[string]map[string]*PlayerInfoClient

	cellMasterMutex      *sync.Mutex
	Cells                *map[string]Cell
	splitCellRequirement int
	splitCheckInterval   int
}

func NewPlayer(splitCellRequirement int, splitCheckInterval int) *Player {
	emptyObjectList := make([]generated.SingleObject, 0)
	cells := make(map[string]Cell, 0)
	emptyPlayerMap := make(map[string]map[string]*PlayerInfoClient, 0)
	mutatedObjects := make([]generated.SingleObject, 0)
	cmConn := CellMasterConnection{}
	mutex := &sync.Mutex{}
	return &Player{
		MutatedObjects:       &mutatedObjects,
		CellMasterConnection: &cmConn,
		SubscribedPlayers:    &emptyPlayerMap,
		MutatingObjects:      &emptyObjectList,
		cellMasterMutex:      mutex,
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
	for _, cell := range in.Cells {
		if ownedCell, ok := (*cm.Cells)[cell.CellId]; ok {
			// cm is already aware of mastership over cell, update cell status
			ownedCell.PosX = cell.PosX
			ownedCell.PosY = cell.PosY
			ownedCell.Height = cell.Height
			ownedCell.Width = cell.Width

			(*cm.Cells)[cell.CellId] = ownedCell
		} else {
			if true {
				println("Received cell mastership with (width, height)", cell.Width, ", ", cell.Height)
			}
			(*cm.Cells)[cell.CellId] = Cell{CellId: cell.CellId, PosX: cell.PosX, PosY: cell.PosY, Width: cell.Width, Height: cell.Height}
			cm.SubscribePlayer(ctx, &generated.PlayerInfo{Port: int32(cm.Port), Ip: cm.Ip, PosY: cm.PosY, PosX: cm.PosX, ObjectId: cm.ObjectId})
		}
	}

	println("Cell list size: ", len(*cm.Cells))

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
				if true {
					println("sending updated objects to player: ", player.Port)
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
//func (cm *Player) GetCellState(ctx context.Context, in *generated.Cell) (*generated.MultipleObjects, error) {
//	return &generated.MultipleObjects{}, nil
//}

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
		println("collideCheck cell posX: ", cell.PosX, " posY: ", cell.PosY, " width: ", cell.Width, " height: ", cell.Height, " Player posX: ", in.PosX, " posY: ", in.PosY)
		if cell.CollidesWith(&cellmanager.Position{PosX: in.PosX, PosY: in.PosY}) {
			if _, exists := (*cm.SubscribedPlayers)[cell.CellId]; !exists {
				(*cm.SubscribedPlayers)[cell.CellId] = map[string]*PlayerInfoClient{}
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
				if true {
					println("Actually subscribing player: ", in.Port)
				}

				subscriberConn := PlayerInfoClient{
					PlayerClient: generated.NewPlayerClient(conn),
					Port:         int(in.Port),
					Ip:           in.Ip,
					ObjectId:     in.ObjectId,
				}
				subscribers[in.Ip+":"+strconv.Itoa(int(in.Port))] = &subscriberConn
			}
			subscribedToCell = true
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
			println("Splitting cell")
			cm.SplitCell(client, cellId)
			cellMap := make(map[string]Cell, 0)
			cm.Cells = &cellMap
		}

		time.Sleep(time.Second * time.Duration(cm.splitCheckInterval))
	}
}

func (cm *Player) SplitCell(client *cellmanager.CellManagerClient, cellID string) {
	cm.cellMasterMutex.Lock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := (*client).DivideCell(ctx, &cellmanager.CellRequest{CellId: cellID})
	if err != nil {
		println("Failed to split cell")
		return
	}

	cm.DesubscribePlayers(ctx)
	// reset subscribed players map
	newSubscribedPlayerMap := make(map[string]map[string]*PlayerInfoClient, 0)
	cm.SubscribedPlayers = &newSubscribedPlayerMap
	cm.cellMasterMutex.Unlock()

}

func (cm *Player) DesubscribePlayers(ctx context.Context) {
	for _, playerMap := range *cm.SubscribedPlayers {
		for _, player := range playerMap {
			(*player).ChangedCellMaster(ctx, &generated.ChangedCellMasterRequest{})
		}
	}
}

func (cm *Player) PlayerIsInOwnedCell(position cellmanager.Position) bool {
	for _, cell := range *cm.Cells {
		if cell.CollidesWith(&position) {
			println("player is out of cell with x: ", position.PosX, " y: ", position.PosY, " and cellX: ", cell.PosX, ", cellY: ", cell.PosY, ", width: ", cell.Width, ", height: ", cell.Height)
			return true
		}
	}
	return false
}

func (cm *Player) PlayerMightLeaveCellHandle(object generated.SingleObject, cellManager *cellmanager.CellManagerClient) {
	//cm.cellMasterMutex.Lock()
	keysAndIndexesToRemove := make(map[string]string, 0)

	println("player might leave cell! with cellid ", object.CellId, " and length: ", len(object.CellId), ", I am responsible for number of cells: ", len(*cm.Cells))

	if len(object.CellId) > 0 {
		return
	}

	for cellId, playerList := range *cm.SubscribedPlayers {
		for playerKey, player := range playerList {
			println("iteratedID: ", player.ObjectId, ", looking for ID: ", object.ObjectId)
			if player.ObjectId == object.ObjectId {

				println("Player left cell, kicking player ", player.Port)
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				player.ChangedCellMaster(ctx, &generated.ChangedCellMasterRequest{})
				if player.ObjectId == cm.ObjectId {
					cm.stopBeingCellMasterForCell(cellManager, cellId)
				}

				keysAndIndexesToRemove[cellId] = playerKey

			}
		}
	}

	for cellKey, playerKey := range keysAndIndexesToRemove {
		delete((*cm.SubscribedPlayers)[cellKey], playerKey)
	}
	//cm.cellMasterMutex.Unlock()
}

func (cm *Player) stopBeingCellMasterForCell(cellManager *cellmanager.CellManagerClient, cellId string) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	(*cellManager).UnregisterCellMaster(ctx, &cellmanager.CellMasterRequest{CellId: cellId})
	delete(*cm.Cells, cellId)
}
