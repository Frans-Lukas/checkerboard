package cellMaster

import (
	"context"
	"fmt"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	OBJ "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
	port = int32(50053)
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NS.NewCellManagerClient(conn)

	// Setup cellMaster server
	lis, err := net.Listen("tcp", ":" + fmt.Sprint(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	cellMaster := objects.NewCellMaster()
	OBJ.RegisterCellMasterServer(s, &cellMaster)

	go func(s grpc.Server) {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}(*s)
}