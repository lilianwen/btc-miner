package node

import (
	"btcnetwork/p2p"
	"net"
	"sync"
	"time"
)

type Peer struct {
	Version       int32
	Conn          net.Conn
	Addr          string    //形式如ip:port
	Alive         chan bool //用于查询节点是否在线
	HandShakeDone chan bool
	SyncBlockDone chan bool
}

func NewPeer() Peer {
	p := Peer{}
	p.Alive = make(chan bool, 1)
	p.HandShakeDone = make(chan bool, 1)
	p.SyncBlockDone = make(chan bool, 1)
	return p
}

func (node *Node) PingPeers(wg *sync.WaitGroup) {
deadloop:
	for {
		select {
		case <-node.PingTicker.C:
			node.Peers.Range(func(key, value interface{}) bool {
				peer := value.(Peer)
				if len(peer.Alive) == 1 { //说明这段时间内接收到该节点的pong消息
					return true
				}

				//给该节点发送ping消息
				msg := p2p.NewPingMsg()
				log.Infof("send ping message to peer[%s]", peer.Addr)
				if err := node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
					log.Error(err)
					return true //todo:把这个节点删除
				}
				return true
			})

			timerCheckAlive := time.NewTimer(time.Duration(node.Cfg.CheckPongTimeval) * time.Second)
			go node.CheckPeerAlive(timerCheckAlive)

		case <-node.StopPing:
			break deadloop
		}
	}

	wg.Done()
}

func (node *Node) CheckPeerAlive(t *time.Timer) {
	//发送完ping消息后等一段时间再遍历一次
	<-t.C

	var deadPeers []string
	node.Peers.Range(func(key, value interface{}) bool {
		addr := key.(string)
		peer := value.(Peer)
		if len(peer.Alive) == 1 { //说明这段时间内接收到该节点的pong消息
			<-peer.Alive
		} else {
			// 可以考虑移除这些被判定为异常的节点,有可能远端节点只是代码有bug，
			deadPeers = append(deadPeers, addr)
		}
		return true
	})

	for _, addr := range deadPeers {
		value, _ := node.Peers.Load(addr)
		peer := value.(Peer)

		if err := peer.Conn.Close(); err != nil {
			log.Error(err)
		}
		node.Peers.Delete(addr)
		log.Infof("peer[%s] is outbound, close the connection.\n", addr)
	}
}
