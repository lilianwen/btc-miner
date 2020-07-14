package common

type Config struct {
	Mine             int
	PingPeriod       int
	CheckPongTimeval int
	MaxPeer          int
	MaxTxInBlock     int
	PeerListen       string
	RpcListen        string
	RemotePeers      []string
	DnsSeed          []string
	DataDir          string
}
