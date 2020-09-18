package v1

import (
	"context"
	pb "github.com/Frans-Lukas/checkerboard/pkg/generated/v1"
)

type CellManager struct {
	pb.CellManagerServer
}

func (s *CellManager) CreateCell (ctx context.Context)