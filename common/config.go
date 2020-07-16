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
	MinerAddr        string
	MinerBanner      string
	MineEmptyBlock   bool
	MineTimeval      int
	FixedTxsInBlock  int
}
