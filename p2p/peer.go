package p2p

import (
	"log"
	"net"
	"sync"
	"time"
)

type Peer struct {
	Version int32
	Conn    net.Conn
	Addr    string    //形式如ip:port
	Alive   chan bool //用于查询节点是否在线
}

func NewPeer() Peer {
	p := Peer{}
	p.Alive = make(chan bool, 1)
	return p
}

// todo:需要考虑共享资源竞争问题
func (node *Node) CheckPeerAlive(wg *sync.WaitGroup) error {
	for {
		time.Sleep(90 * time.Second) //todo: 从配置文件里读出来

		for _, peer := range node.Peers {
			if len(peer.Alive) == 1 { //说明这段时间内接收到该节点的pong消息
				continue
			}

			//给该节点发送ping消息
			msg := NewPingMsg()

			if err := node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
				log.Println(err)
				continue
			}
		}

		time.Sleep(5 * time.Second) //todo: 从配置文件里读出来
		//发送完ping消息后等一段时间再遍历一次
		node.mu.Lock()
		for _, peer := range node.Peers {
			if len(peer.Alive) == 1 { //说明这段时间内接收到该节点的pong消息
				<-peer.Alive
			} else {
				// 可以考虑移除这些被判定为异常的节点,有可能远端节点只是代码有bug，没有及时回pong消息
				peer.Conn.Close()
				delete(node.Peers, peer.Conn.RemoteAddr().String())
			}
		}
		node.mu.Unlock()
	}

	wg.Done()
	return nil
}
