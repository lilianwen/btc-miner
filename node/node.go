package node

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"btcnetwork/p2p"
	"btcnetwork/storage"
	"encoding/hex"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
	//"internal/poll"
)

const (
	MaxSyncPeerCount = 3
)

//需要种子节点列表
type Node struct {
	Cfg            common.Config                               //从配置文件读出来以后就不会再被改变了
	Handlers       map[string]func(*Node, *Peer, []byte) error //存储各个消息的处理函数
	Peers          map[string]Peer                             //按照地址映射远程节点的信息
	PeerAmount     uint32
	mu             sync.RWMutex
	PingTicker     *time.Ticker
	StopPing       chan bool
	txPool         map[[32]byte][]byte
	syncBlocksDone chan bool
	exit           bool
}

var defaultNode *Node

func Start(cfg *common.Config) {
	defaultNode = newNode(cfg) //写成new会和golang内置的new重名
	go defaultNode.Start()
}

func Stop() {
	defaultNode.exit = true
	// 关闭定时检测节点服务
	defaultNode.closePeerCheck()
	// 关闭所有的p2p连接节点
	defaultNode.disconnetAllPeers()
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
		msg, _ := p2p.NewVerMsg(addr)

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
		if !node.exit {
			wg.Add(1)
			go node.SyncBlock(conn, &wg)
			// 等待区块同步完成
			<-node.syncBlocksDone
		}

		if !node.exit {
			wg.Add(1)
			go node.SyncMempool(&wg)
		}

		if !node.exit {
			wg.Add(1)
			go node.StartApiService(&wg)
		}
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

func newNode(cfg *common.Config) *Node {
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
	var syncBlocksDone = make(chan bool, 1)
	for _, addr := range cfg.RemotePeers {
		peers[addr] = NewPeer()
	}
	return &Node{Cfg: *cfg, Peers: peers, Handlers: handlers, txPool: txpool, syncBlocksDone: syncBlocksDone, exit: false}
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

func readMsgHeader(conn net.Conn) (p2p.MsgHeader, error) {
	var sum = 0
	var start = 0
	var buf = make([]byte, p2p.MsgHeaderLen) //
	for sum < p2p.MsgHeaderLen {
		//先读取消息头,再读payload
		n, err := conn.Read(buf[start:]) //万一多读了数据怎么办？
		if err != nil {
			return p2p.MsgHeader{}, err
		}
		sum += n
		start = sum
	}
	header := p2p.MsgHeader{}
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
				log.Infof("remote peer(%s) close connection.", conn.RemoteAddr().String())
			} else {
				log.Errorln(err) // todo: 解决ERRO[0006] read tcp 127.0.0.1:57380->127.0.0.1:9333: use of closed network connection
			}

			break
		}

		cmd := common.Byte2String(header.Command[:])
		log.Printf("received from [%s] message:[%s]", conn.RemoteAddr().String(), cmd)

		payload, err := readPayload(conn, header.LenOfPayload)
		if err != nil {
			if err == io.EOF {
				log.Infof("remote peer(%s) close connection.", conn.RemoteAddr().String())
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

func (node *Node) closePeerCheck() {
	//close(node.PingTicker)
	node.PingTicker.Stop()
}

func (node *Node) disconnetAllPeers() {
	for _, peer := range node.Peers {
		_ = peer.Conn.Close()
	}
}

func (node *Node) HandleVersion(peer *Peer, payload []byte) error {
	//根据网络协议，收到version消息，就应该发送一个verack报文给对方
	versionPayload := p2p.VersionPayload{}
	versionPayload.Parse(payload)
	peer.Version = versionPayload.Version

	verackMsg, err := p2p.NewMsg("verack", nil)
	if err != nil {
		return err
	}
	if err = p2p.MustWrite(peer.Conn, verackMsg.Serialize()); err != nil {
		return err
	}

	return nil
}

func (node *Node) HandleVerack(peer *Peer, payload []byte) error {
	if len(peer.HandShakeDone) == 0 {
		peer.HandShakeDone <- true
	}

	return nil
}

func (node *Node) HandleBlock(peer *Peer, payload []byte) error {
	log.Infof("receive block data from [%s]", peer.Addr)

	recvBlock := p2p.BlockPayload{}
	err := recvBlock.Parse(payload)
	if err != nil {
		return err
	}

	//开始验证区块的合法性
	blockHash := common.Sha256AfterSha256(payload[:recvBlock.Header.Len()])
	log.Infoln("block hash:", hex.EncodeToString(blockHash[:]))
	log.Infoln("block merkle root hash:", hex.EncodeToString(recvBlock.MerkleRootHash[:]))
	log.Infoln("pre block hash:", hex.EncodeToString(recvBlock.PreHash[:]))
	//处理区块里的交易数据

	var hashes []string
	for i := range recvBlock.Txns {
		txHash := recvBlock.Txns[i].TxHash()
		//删除mempool中的交易
		node.mu.Lock()
		if _, ok := node.txPool[txHash]; ok {
			delete(node.txPool, txHash)
		}
		node.mu.Unlock()

		//strHash := hex.EncodeToString(common.ReverseBytes(txHash[:]))
		log.Infof("tx[%d].txhash=%s, size=%d", i, hex.EncodeToString(txHash[:]), recvBlock.Txns[i].Len())
		txid := recvBlock.Txns[i].Txid()
		strID := hex.EncodeToString(common.ReverseBytes(txid[:]))
		log.Infof("tx[%d].txid=%s", i, strID)
		hashes = append(hashes, strID)

	}
	//构建默克尔树验证区块头的默克尔根值
	root, err := block.ConstructMerkleRoot(hashes)
	wantMerkleRootHash := hex.EncodeToString(recvBlock.MerkleRootHash[:])
	if root.Value != wantMerkleRootHash {
		log.Error("calculate merkle root hash not equal to block header merkle root hash")
		log.Errorf("get:%s, want:%s", root.Value, wantMerkleRootHash)
	}

	// 把收到的区块数据存进leveldb
	// storage.Store(&recvBlock)
	return nil
}

func (node *Node) SyncBlock(conn net.Conn, wg *sync.WaitGroup) {
	//发送getblocks消息给远程节点
	payload := p2p.GetblocksPayload{}
	payload.Version = uint32(70002)                                               //todo:各种version版本要灾难了，这里需要总结一下，不然乱了
	payload.HashCount = common.NewVarInt(uint64(storage.LatestBlockHeight() + 1)) //表示我目前有多少个区块
	latestBlockHash := storage.LatestBlockHash()
	copy(payload.HashStart[:], latestBlockHash[:])
	msg, err := p2p.NewMsg("getblocks", payload.Serialize())
	if err != nil {
		log.Error(err)
		panic(err)
	}
	if err = p2p.MustWrite(conn, msg.Serialize()); err != nil {
		log.Error(err)
		panic(err)
	}
	log.Infof("sync block from [%s]...", conn.RemoteAddr().String())
	wg.Done()
}

func (node *Node) HandleGetblocks(peer *Peer, payload []byte) error {
	gp := p2p.GetblocksPayload{}
	err := gp.Parse(payload)
	if err != nil {
		return err
	}
	log.Println("get hash count:", gp.HashCount.Value)
	log.Println("hash start:", hex.EncodeToString(gp.HashStart[:]))
	log.Println("hash stop:", hex.EncodeToString(gp.HashStop[:]))

	//对方节点告诉我他的最新的区块hash值是多少，他想要的最新的区块哈希值是多少（如果是全0）表示他想要我的最新的区块
	//我对比一下我这边的最新区块哈希值和它的是不是相同，如果相同表示我们拥有相同的区块数据，
	//如果不同，则我的更旧，我就向它要区块（发送getblocks消息），如果我的更新，我就发送inv消息给他，向它提供最新的区块
	latestBlockHash := storage.LatestBlockHash()
	if hex.EncodeToString(latestBlockHash[:]) != hex.EncodeToString(gp.HashStart[:]) {
		if storage.HasBlockHash(gp.HashStart) { //说明我的区块较对方节点更新
			// todo:把我的区块发送给他
			return errors.New("not support sending blocks to remote peers")
		} else {
			// todo:向对方请求最新的区块
			return errors.New("not support receiving blocks from remote peers")
		}

	}

	//发送inv消息表明我拥有的区块和对方一样
	invp := p2p.InvPayload{}
	invp.Count = common.NewVarInt(0)
	msg, err := p2p.NewMsg("inv", invp.Serialize())
	if err != nil {
		return err
	}
	if err = node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
		return err
	}

	return nil
}

//
//func (node *Node) GetLatestBlockHash() string {
//	return "f67ad7695d9b662a72ff3d8edbbb2de0bfa67b13974bb9910d116d5cbd863e68"
//}

func (node *Node) HandleGetdata(peer *Peer, payload []byte) error {
	//主要处理block和tx这两种数据，其他的暂时忽略
	ivp := p2p.InvPayload{}
	err := ivp.Parse(payload)
	if err != nil {
		return err
	}
	for i := uint64(0); i < ivp.Count.Value; i++ {
		switch ivp.Inventory[i].Type {
		case common.MsgTx:
			{
				//todo:发送tx消息给对方

			}
		case common.MsgBlock:
			{
				//todo:发送block消息给对方

			}
		case common.MsgFilteredBlock:
			fallthrough
		case common.MsgCmpctBlock:
			fallthrough
		case common.MsgErr:
			fallthrough
		default:
			//什么都不做，忽略
		}
	}

	return nil
}

func (node *Node) HandleInv(peer *Peer, payload []byte) error {
	//主要处理block和tx这两种数据，其他的暂时忽略
	ivp := p2p.InvPayload{}
	err := ivp.Parse(payload)
	if err != nil {
		return err
	}
	var count = uint64(0)
	var toGetData = false
	var invVect []common.InvVector
	for i := uint64(0); i < ivp.Count.Value; i++ {
		toGetData = false
		hash := ivp.Inventory[i].Hash
		switch ivp.Inventory[i].Type {
		case common.MsgTx:
			{
				//存储到交易池中去,交易池用什么存？leveldb？rockdb?
				//查询本地是否有该交易？没有就验证交易，通过验证之后就加入交易池
				//todo:暂时省去验证交易的功能，直接认为是合法交易，将来再加上
				if _, ok := node.txPool[hash]; !ok {
					//本地没有这个tx,发送getdata消息给节点获取交易数据
					//如果不想知道交易详情，其实是可以不去get交易数据的
					toGetData = true
				}

			}
		case common.MsgBlock:
			{
				//存储到区块中去,区块怎么存储到硬盘上？用什么格式？要不要压缩？
				//查看本地是否有该区块，如果没有就验证区块和区块里的交易，通过之后就加入本地区块
				if !HasBlock(hash) {
					//本地没有这个tx,发送getdata消息给节点获取交易数据
					toGetData = true
				}
			}
		case common.MsgFilteredBlock:
			fallthrough
		case common.MsgCmpctBlock:
			fallthrough
		case common.MsgErr:
			fallthrough
		default:
			//什么都不做，忽略
		}
		if toGetData {
			count++
			invVect = append(invVect, ivp.Inventory[i])
		}
	}
	if count != uint64(0) {
		//发送getdata消息给节点获取数据
		// 这里要注意，getdata每次最多发送128条数据条目
		msg, err := p2p.NewMsg("getdata", p2p.NewGetdataPayload(count, invVect).Serialize())
		if err != nil {
			return err
		}
		if err = node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
			return err
		}
	}

	return nil
}

//todo:查看本地是否有这个区块哈希对应的区块，有就返回true.暂时不知道放哪，先放这里好了。
func HasBlock(blockHash [32]byte) bool {
	_ = blockHash
	return false
}

//todo:这里需要好好想想，怎样才算是真正的同步成功了呢？
func (node *Node) SyncMempool(wg *sync.WaitGroup) {
	//发送mempool消息，接收inv消息
	node.mu.RLock()
	count := 0
	for addr, peer := range node.Peers {
		msg, err := p2p.NewMempoolMsg()
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
	node.mu.RUnlock()
	wg.Done()
}

func (node *Node) HandlePing(peer *Peer, payload []byte) error {
	var err error

	msgPong, err := p2p.NewMsg("pong", payload)
	if err != nil {
		return err
	}
	if err = p2p.MustWrite(peer.Conn, msgPong.Serialize()); err != nil {
		return err
	}
	return nil
}

func (node *Node) HandlePong(peer *Peer, payload []byte) error {
	//实现心跳机制
	if len(peer.Alive) == 0 { //可能远程节点发送ping更频繁一些，这里做这个判断防止通道满了再往里写会阻塞
		node.mu.Lock()
		peer.Alive <- true
		node.mu.Unlock()
	}

	return nil
}

func (node *Node) HandleTx(peer *Peer, payload []byte) error {
	_ = peer
	//todo:验证交易是否合法
	//如果交易合法，则查看本地是否有该交易，如果没有就加入本地交易池
	txHash := common.Sha256AfterSha256(payload)

	//看来还挺难搞哦，这个可能刚刚被打包进区块了，所以交易池没有这个交易，如果就这么再添加到交易池的话，那就是很大的bug了。
	//要解决这种问题，看来只能把所有区块都同步下来并回，才能确定唯一性。
	node.mu.Lock()
	if _, ok := node.txPool[txHash]; !ok { //todo:先暂时这么粗暴的做，这里肯定有bug
		node.txPool[txHash] = payload
	}
	node.mu.Unlock()
	return nil
}

var log *logrus.Logger

func init() {
	log = logrus.New()
}
