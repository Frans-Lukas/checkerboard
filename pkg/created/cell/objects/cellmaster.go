package objects

import (
	"fmt"
)

func ToAddress(ip string, port int32) string {
	return ip + ":" + fmt.Sprint(port)
}
