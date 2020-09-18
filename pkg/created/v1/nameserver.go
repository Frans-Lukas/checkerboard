package v1

import (
	"context"
	pb "github.com/Frans-Lukas/checkerboard/pkg/generated/v1"
)

type CellManager struct {
	pb.CellManagerServer
}

func (cellManager *CellManager) CreateCell(ctx context.Context, in *pb.CellRequest) (*pb.CellStatusReply, error) {
	return &pb.CellStatusReply{WasPerformed: true}, nil
}
