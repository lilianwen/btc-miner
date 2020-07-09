package p2p

import "sync"

const (
	MaxSyncPeerCount = 3
)

func NewMempoolMsg() (*Msg, error) {
	return NewMsg("mempool", nil)
}

//todo:这里需要好好想想，怎样才算是真正的同步成功了呢？还是直接放任不管了？
func (node *Node) SyncMempool(wg *sync.WaitGroup) {
	//发送mempool消息，接收inv消息
	node.mu.Lock()
	count := 0
	for addr, peer := range node.Peers {
		msg, err := NewMempoolMsg()
		if err != nil {
			panic(err) //如果连消息都会构造错误那只能panic了
		}
		if err = node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
			log.Errorln(err)
			continue
		}
		log.Infof("sync memery pool from [%s]", addr)

		count++
		if count >= MaxSyncPeerCount {
			break
		}
	}
	node.mu.Unlock()
	wg.Done()
}
