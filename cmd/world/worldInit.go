package main

import (
	"context"
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
	port        = int32(50053)
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NS.NewCellManagerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.SetWorldSize(ctx, &NS.WorldSize{Height: constants.MAP_SIZE, Width: constants.MAP_SIZE})
}
