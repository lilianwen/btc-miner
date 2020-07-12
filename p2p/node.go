package p2p

import (
	"btcnetwork/common"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

//需要种子节点列表
type Node struct {
	Cfg        common.Config                               //从配置文件读出来以后就不会再被改变了
	Handlers   map[string]func(*Node, *Peer, []byte) error //存储各个消息的处理函数
	Peers      map[string]Peer                             //按照地址映射远程节点的信息
	PeerAmount uint32
	mu         sync.RWMutex
	PingTicker *time.Ticker
	StopPing   chan bool
	txPool     map[[32]byte][]byte
}

func (node *Node) Start() {
	var wg sync.WaitGroup
	//todo:先找到本地的区块数据和交易数据，然后回放到内存中去

	//主动连接P2P节点
	for addr, peer := range node.Peers {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Errorln(err.Error())
			continue
		}
		//开始握手协议
		//既然前面能够拨号成功，那么ip和端口肯定是合法的，就不用检查错误了
		msg, _ := NewVerMsg(addr)

		if err = node.sendMsg(conn, msg.Serialize()); err != nil {
			log.Errorln(err.Error())
			continue
		}
		peer.Conn = conn
		peer.Addr = addr
		node.AddPeer(peer)

		wg.Add(1)
		go node.handleMsg(conn, &wg)

		//time.Duration()
		node.PingTicker = time.NewTicker(time.Duration(node.Cfg.PingPeriod) * time.Second)
		node.StopPing = make(chan bool)
		wg.Add(1)
		go node.PingPeers(&wg)

		//一定要注意时序，必须要在握手协议完成之后才能进行后续的P2P通信
		<-peer.HandShakeDone
		wg.Add(1)
		go node.SyncMempool(&wg)

		//
		wg.Add(1)
		go node.StartApiService(&wg)
	}

	go node.listenPeers(&wg)

	wg.Wait()
}

func (node *Node) listenPeers(wg *sync.WaitGroup) {
	defer wg.Done()

	listener, err := net.Listen("tcp", node.Cfg.PeerListen)
	if err != nil {
		log.Errorln(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorln(err)
			continue
		}
		//节点数量有个上限，不能一直在这里接收
		if atomic.LoadUint32(&node.PeerAmount) >= uint32(node.Cfg.MaxPeer) {
			conn.Close()
			continue
		}

		p := NewPeer()
		p.Conn = conn
		p.Addr = conn.RemoteAddr().String()
		node.AddPeer(p)

		go node.handleMsg(conn, wg)
	}
}

func NewNode(cfg *common.Config) *Node {
	var handlers = map[string]func(*Node, *Peer, []byte) error{
		"version":   (*Node).HandleVersion,
		"verack":    (*Node).HandleVerack,
		"ping":      (*Node).HandlePing,
		"pong":      (*Node).HandlePong,
		"getblocks": (*Node).HandleGetblocks,
		"inv":       (*Node).HandleInv,
		"block":     (*Node).HandleBlock,
		"tx":        (*Node).HandleTx,
	}
	var peers = make(map[string]Peer)
	var txpool = make(map[[32]byte][]byte)
	for _, addr := range cfg.RemotePeers {
		peers[addr] = NewPeer()
	}
	return &Node{Cfg: *cfg, Peers: peers, Handlers: handlers, txPool: txpool}
}

// 新增节点，主要给监听服务和addr消息用
func (node *Node) AddPeer(p Peer) {
	node.mu.Lock()
	node.Peers[p.Addr] = p
	node.mu.Unlock()
	atomic.AddUint32(&node.PeerAmount, 1)
}

func (node *Node) sendMsg(conn net.Conn, data []byte) error {
	var sum = 0
	var start = 0
	for sum < len(data) { //防止少发送数据
		n, err := conn.Write(data[start:])
		if err != nil {
			return err
		}
		sum += n
		start = sum
	}

	return nil
}

func readMsgHeader(conn net.Conn) (MsgHeader, error) {
	var sum = 0
	var start = 0
	var buf = make([]byte, MsgHeaderLen) //
	for sum < MsgHeaderLen {
		//先读取消息头,再读payload
		n, err := conn.Read(buf[start:]) //万一多读了数据怎么办？
		if err != nil {
			return MsgHeader{}, err
		}
		sum += n
		start = sum
	}
	header := MsgHeader{}
	header.Parse(buf)
	return header, nil
}

func readPayload(conn net.Conn, payloadLen uint32) ([]byte, error) {
	var sum = 0
	var start = 0
	var buf = make([]byte, payloadLen) //
	for sum < int(payloadLen) {
		//先读取消息头,再读payload
		n, err := conn.Read(buf[start:])
		if err != nil {
			return nil, err
		}
		sum += n
		start = sum
	}
	return buf, nil
}

func (node *Node) handleMsg(conn net.Conn, wg *sync.WaitGroup) {
	for {
		header, err := readMsgHeader(conn)
		if err != nil {
			if err == io.EOF {
				log.Errorf("remote peer(%s) close connection.", conn.RemoteAddr().String())
			} else {
				log.Errorln(err.Error())
			}

			break
		}

		cmd := common.Byte2String(header.Command[:])
		log.Printf("received from [%s] message:[%s]", conn.RemoteAddr().String(), cmd)

		payload, err := readPayload(conn, header.LenOfPayload)
		if err != nil {
			if err == io.EOF {
				log.Errorf("remote peer(%s) close connection.", conn.RemoteAddr().String())
			} else {
				log.Errorln(err)
			}

			break
		}

		handler, ok := node.Handlers[cmd]
		if !ok {
			log.Errorf("not support message(%s) handler", cmd)
			continue
		}
		peer := node.Peers[conn.RemoteAddr().String()]
		if err = handler(node, &peer, payload); err != nil {
			log.Errorf("handle message(%s) error:%s", cmd, err.Error())
			break
		}
	}

	//释放资源
	conn.Close()
	//删除节点
	node.mu.Lock()
	delete(node.Peers, conn.RemoteAddr().String())
	node.mu.Unlock()
	wg.Done()
}

var log *logrus.Logger

func init() {
	log = logrus.New()
	//log.
}
