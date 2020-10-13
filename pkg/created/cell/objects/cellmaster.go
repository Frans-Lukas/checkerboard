package objects

import (
	"fmt"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
)

func CMToAddress(reply cellmanager.CellMasterReply) string {
	return reply.Ip + ":" + fmt.Sprint(reply.Port)
}
