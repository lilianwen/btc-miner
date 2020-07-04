package p2p

import "net"

type Peer struct {
	Version int32
	Conn    net.Conn
	Addr    string //形式如ip:port
}
