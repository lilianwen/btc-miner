package p2p

import (
	"btcnetwork/common"
	"io"
	"log"
	"net"
)

//需要种子节点列表
type Node struct {
	Handlers map[string]func(*Node, *Peer, []byte) error //存储各个消息的处理函数
	Peers    map[string]Peer                             //按照地址映射远程节点的信息
}

func (node *Node) Start() {
	//主动连接P2P节点
	for addr, peer := range node.Peers {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		//开始握手协议
		//既然前面能够拨号成功，那么ip和端口肯定是合法的，就不用检查错误了
		msg, _ := NewVerMsg(addr)

		if err = node.sendMsg(conn, msg.Serialize()); err != nil {
			log.Println(err.Error())
			continue
		}
		peer.Conn = conn
		node.Peers[addr] = peer

		go node.handleMsg(conn)

		go node.CheckPeerAlive()
	}

	//todo:启动监听服务准备接受其他节点的连接
	listener, err := net.Listen("tcp", "0.0.0.0:8333")
	if err != nil {
		log.Panicln(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panicln(err)
		}
		node.AddPeer(conn, conn.RemoteAddr().String())
		go node.handleMsg(conn)
	}
}

func NewNode(addresses []string) *Node {
	var handlers = map[string]func(*Node, *Peer, []byte) error{
		"version":    (*Node).HandleVersion,
		"verack": (*Node).HandleVerack,
		"ping":   (*Node).HandlePing,
		"pong":   (*Node).HandlePong,
		"getblocks": (*Node).HandleGetblocks,
	}
	var mapPeers = make(map[string]Peer)
	for _, addr := range addresses {
		mapPeers[addr] = NewPeer()
	}
	return &Node{Peers: mapPeers, Handlers: handlers}
}

// 新增节点，主要给监听服务和addr消息用
func (node *Node) AddPeer(conn net.Conn, addr string) {
	p := NewPeer()
	p.Conn = conn
	p.Addr = addr
	node.Peers[addr] = p
}

func (node *Node)sendMsg(conn net.Conn, data []byte) error {
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

func (node *Node) handleMsg(conn net.Conn) {
	for {
		header, err := readMsgHeader(conn)
		if err != nil {
			if err == io.EOF {
				log.Printf("remote peer(%s) close connection.", conn.RemoteAddr().String())
			} else {
				log.Println(err.Error())
			}

			break
		}

		log.Printf("received from %s message:%s\n", conn.RemoteAddr().String(), string(header.Command[:]))

		payload, err := readPayload(conn, header.LenOfPayload)
		if err != nil {
			if err == io.EOF {
				log.Printf("remote peer(%s) close connection.", conn.RemoteAddr().String())
			} else {
				log.Println(err)
			}

			break
		}
		cmd := common.Byte2String(header.Command[:])
		handler, ok := node.Handlers[cmd]
		if !ok {
			log.Printf("not support message(%s) handler\n", cmd)
			continue
		}
		peer := node.Peers[conn.RemoteAddr().String()]
		if err = handler(node, &peer, payload); err != nil {
			log.Printf("handle message(%s) error:%s\n", cmd, err.Error())
			break
		}
	}

	//释放资源
	conn.Close()
	//删除节点
	delete(node.Peers, conn.RemoteAddr().String())
}
