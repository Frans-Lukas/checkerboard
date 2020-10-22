package main

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	"google.golang.org/grpc"
	"log"
	"os"
	"time"
)

func main() {
	//1. connect to cell manager

	conn, err := grpc.Dial(constants.CellManagerAddress, grpc.WithInsecure(), grpc.WithTimeout(time.Millisecond*constants.DialTimeoutMilli))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NS.NewCellManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c.AddPlayerToCellWithPositions(ctx, &NS.PlayerInCellRequestWithPositions{Ip: os.Args[1], Port: int32(port), PosX: 0, PosY: 0})

	cm, err := c.RequestCellMasterWithPositions(ctx, &NS.Position{PosX: 0, PosY: 0})

	//2. get reference to all cell masters/ cells.



	//3. continuously poll cell masters for objects to update.
	//3,5. check if previous reference to object exists.
	//4. apply game logic to updates.
}
