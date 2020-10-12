package objects

import (
	"fmt"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
)

func ToAddress(reply cellmanager.CellMasterReply) string {
	return reply.Ip + ":" + fmt.Sprint(reply.Port)
}
