package objects

type Player struct {
	Port int32
	Ip   string
	// 0 = lowest trust level UINT32_MAX = highest trust level
	TrustLevel uint32
}

