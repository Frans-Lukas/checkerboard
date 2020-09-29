package objects

import (
	"context"
	"errors"
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
	CellMaster      Client
	ObjectsToUpdate map[string]map[string]string
}

func NewPlayer() Player {
	return Player{CellMaster: Client{Port: -1, Ip: "none"}, ObjectsToUpdate: map[string]map[string]string{}}
}

func (player *Player) UpdateCellMaster(
	ctx context.Context, in *generated.NewCellMaster,
) (*generated.EmptyReply, error) {
	player.CellMaster = Client{Ip: in.Ip, Port: in.Port}
	return &generated.EmptyReply{}, nil
}

func (player *Player) SendUpdate(
	ctx context.Context, in *generated.MultipleObjects,
) (*generated.EmptyReply, error) {
	for _, object := range in.Objects {
		if len(object.NewValue) != len(object.UpdateKey) {
			return &generated.EmptyReply{}, errors.New("not as many values as keys")
		}
	}

	for _, object := range in.Objects {

		if player.ObjectsToUpdate[object.ObjectId] == nil {
			player.ObjectsToUpdate[object.ObjectId] = map[string]string{}
		}

		objectsToUpdate := player.ObjectsToUpdate[object.ObjectId]
		for index, key := range object.UpdateKey {
			objectsToUpdate[key] = object.NewValue[index]
		}
	}

	return &generated.EmptyReply{}, nil
}
