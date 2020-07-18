package node

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"btcnetwork/p2p"
	"btcnetwork/storage"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MaxSyncPeerCount = 3
	MaxSyncChanLen   = 3
)

//需要种子节点列表
type Node struct {
	Cfg            common.Config                                              //从配置文件读出来以后就不会再被改变了
	p2pHandlers    map[string]func(*Node, *Peer, []byte) error                //存储各个p2p消息的处理函数
	apiHandlers    map[string]func(*Node, http.ResponseWriter, *http.Request) //存储各个api消息的处理函数
	Peers          sync.Map                                                   //按照地址映射远程节点的信息
	PeerAmount     uint32
	PingTicker     *time.Ticker
	StopPing       chan bool
	txPool         sync.Map
	syncBlocksDone chan bool
	syncingBlocks  chan bool
	isSyncBlock    bool
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

	//主动连接P2P节点
	node.Peers.Range(func(key, value interface{}) bool {
		addr := key.(string)
		peer := value.(Peer)

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Error(err.Error())
			return false
		}
		//开始握手协议
		//既然前面能够拨号成功，那么ip和端口肯定是合法的，就不用检查错误了
		msg, _ := p2p.NewVerMsg(addr)

		if err = node.sendMsg(conn, msg.Serialize()); err != nil {
			log.Error(err.Error())
			return false
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
		close(peer.HandShakeDone)
		if !node.exit {
			go node.SyncBlock(conn) //暂时只从一个节点同步区块，todo:将来需要扩展成从多个节点同步区块
			// 等待区块同步完成
			<-node.syncBlocksDone
			node.isSyncBlock = false
			close(node.syncBlocksDone)
			close(node.syncingBlocks)
			log.Info("sync block done.")
		}

		if !node.exit {
			wg.Add(1)
			go node.SyncMempool(&wg)
		}

		if !node.exit {
			wg.Add(1)
			go node.StartApiService(&wg)
		}

		return true
	})

	go node.listenPeers(&wg)

	wg.Wait()
}

func (node *Node) listenPeers(wg *sync.WaitGroup) {
	defer wg.Done()

	listener, err := net.Listen("tcp", node.Cfg.PeerListen)
	if err != nil {
		log.Error(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err)
			continue
		}
		//节点数量有个上限，不能一直在这里接收
		if atomic.LoadUint32(&node.PeerAmount) >= uint32(node.Cfg.MaxPeer) {
			if err = conn.Close(); err != nil {
				log.Error(err)
			}
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
		"getdata":   (*Node).HandleGetdata,
	}

	var syncBlocksDone = make(chan bool, 1)
	var peers sync.Map
	for _, addr := range cfg.RemotePeers {
		peers.Store(addr, NewPeer())
	}
	return &Node{Cfg: *cfg, Peers: peers, p2pHandlers: handlers, syncBlocksDone: syncBlocksDone, isSyncBlock: false, exit: false}
}

// 新增节点，主要给监听服务和addr消息用
func (node *Node) AddPeer(p Peer) {
	node.Peers.Store(p.Addr, p)
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
	header := p2p.MsgHeader{}
	var buf = make([]byte, p2p.MsgHeaderLen)
	if n, err := io.ReadFull(conn, buf); err != nil {
		log.Errorf("read %d bytes when error(%v) occurs", n, err)
		return header, err
	}
	header.Parse(buf)
	return header, nil
}

func readPayload(conn net.Conn, payloadLen uint32) ([]byte, error) {
	var buf = make([]byte, payloadLen)
	if n, err := io.ReadFull(conn, buf); err != nil {
		log.Errorf("read %d bytes when error(%v) occurs", n, err)
		return buf, err
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
				log.Error(err) // todo: 解决ERRO[0006] read tcp 127.0.0.1:57380->127.0.0.1:9333: use of closed network connection
			}

			break
		}

		cmd := common.Byte2String(header.Command[:])
		log.Infof("received from [%s] message:[%s]", conn.RemoteAddr().String(), cmd)

		payload, err := readPayload(conn, header.LenOfPayload)
		if err != nil {
			if err == io.EOF {
				log.Infof("remote peer(%s) close connection.", conn.RemoteAddr().String())
			} else {
				log.Error(err)
			}
			break
		}

		handler, ok := node.p2pHandlers[cmd]
		if !ok {
			log.Errorf("not support message(%s) handler", cmd)
			continue
		}
		addr := conn.RemoteAddr().String()
		value, ok := node.Peers.Load(addr)
		peer := value.(Peer)
		if !ok {
			log.Errorf("peer(%s) not exist anymore.", addr)
			break
		}
		if err = handler(node, &peer, payload); err != nil {
			log.Errorf("handle message(%s) error:%s", cmd, err.Error())
			break
		}
	}

	//释放资源
	if err := conn.Close(); err != nil {
		log.Error(err)
	}
	//删除节点
	node.Peers.Delete(conn.RemoteAddr().String())

	wg.Done()
}

func (node *Node) closePeerCheck() {
	node.PingTicker.Stop()
}

func (node *Node) disconnetAllPeers() {
	node.Peers.Range(func(key, value interface{}) bool {
		_ = key
		peer := value.(Peer)
		if err := peer.Conn.Close(); err != nil {
			log.Error("err")
		}
		return true
	})
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

func (node *Node) removeTxsInBlock(blk *p2p.BlockPayload) []string {
	var hashes []string
	for i := range blk.Txns {
		txHash := blk.Txns[i].TxHash()
		//删除mempool中的交易
		node.txPool.Delete(txHash)

		log.Debug("tx[%d].txhash=%s, size=%d", i, hex.EncodeToString(txHash[:]), blk.Txns[i].Len())
		txid := blk.Txns[i].Txid()
		strID := hex.EncodeToString(common.ReverseBytes(txid[:]))
		log.Debug("tx[%d].txid=%s", i, strID)
		hashes = append(hashes, strID)
	}
	return hashes
}

func (node *Node) HandleBlock(peer *Peer, payload []byte) error {
	log.Infof("receive block data from [%s]", peer.Addr)
	if node.isSyncBlock {
		if len(node.syncingBlocks) < MaxSyncChanLen {
			node.syncingBlocks <- true
		}
	}

	recvBlock := p2p.BlockPayload{}
	err := recvBlock.Parse(payload)
	if err != nil {
		return err
	}

	//开始验证区块的合法性
	log.Debug("blockheader: >", hex.EncodeToString(payload[:recvBlock.Header.Len()]))
	blockHash := common.Sha256AfterSha256(payload[:recvBlock.Header.Len()])
	log.Info("----------------block hash:", hex.EncodeToString(blockHash[:]))
	log.Debug("block merkle root hash:", hex.EncodeToString(recvBlock.MerkleRootHash[:]))
	log.Debug("pre block hash:", hex.EncodeToString(recvBlock.PreHash[:]))
	//处理区块里的交易数据
	txids := node.removeTxsInBlock(&recvBlock)

	//构建默克尔树验证区块头的默克尔根值
	root, err := block.ConstructMerkleRoot(txids) //这里算出来的是给人类看的逆序的merkle根
	var merkleHashCopy [32]byte
	copy(merkleHashCopy[:], recvBlock.MerkleRootHash[:])
	wantMerkleRootHash := hex.EncodeToString(common.ReverseBytes(merkleHashCopy[:]))
	if root.Value != wantMerkleRootHash {
		log.Error("calculate merkle root hash not equal to block header merkle root hash")
		log.Errorf("get:%s, want:%s", root.Value, wantMerkleRootHash)
	}

	// 把收到的区块数据存进leveldb
	storage.Store(&recvBlock)
	return nil
}

func (node *Node) SyncBlock(conn net.Conn) {
	//发送getblocks消息给远程节点
	node.isSyncBlock = true
	node.syncingBlocks = make(chan bool, MaxSyncChanLen) //预留足够的空间，预留足够的检测时间
	payload := p2p.GetblocksPayload{}
	payload.Version = uint32(70002)                 //todo:各种version版本要灾难了，这里需要总结一下，不然乱了
	payload.HashCount = common.NewVarInt(uint64(1)) //注意，这里固定填写1，否则远程节点会直接断开连接
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

	go node.SyncBlockMonitor()
	log.Infof("sync block from [%s]...", conn.RemoteAddr().String())
}

func (node *Node) SyncBlockMonitor() {
	t := time.NewTicker(3 * time.Second)
	for {
		<-t.C
		if len(node.syncingBlocks) == 0 {
			t.Stop()
			break
		}
		_ = <-node.syncingBlocks
	}
	node.syncBlocksDone <- true
}

func (node *Node) HandleGetblocks(peer *Peer, payload []byte) error {
	gp := p2p.GetblocksPayload{}
	err := gp.Parse(payload)
	if err != nil {
		return err
	}
	log.Debug("get hash count:", gp.HashCount.Value)
	log.Debug("hash start:", hex.EncodeToString(gp.HashStart[:]))
	log.Debug("hash stop:", hex.EncodeToString(gp.HashStop[:]))

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
			//todo:发送tx消息给对方

		case common.MsgCmpctAndBlock:
			fallthrough
		case common.MsgBlock:
			blk, err := storage.BlockFromHash(ivp.Inventory[i].Hash)
			if err != nil {
				// 发送reject消息
				log.Error(err)
				return err
			}
			msg, err := p2p.NewMsg("block", blk.Serialize())
			if err != nil {
				// 发送 reject消息
				log.Error(err)
				return err
			}
			if err = p2p.MustWrite(peer.Conn, msg.Serialize()); err != nil {
				log.Error(err)
				return err
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

				if _, ok := node.txPool.Load(hash); !ok {
					//本地没有这个tx,发送getdata消息给节点获取交易数据
					//如果不想知道交易详情，其实是可以不去get交易数据的
					toGetData = true
				}

			}
		case common.MsgBlock:
			{
				//存储到区块中去,区块怎么存储到硬盘上？用什么格式？要不要压缩？
				//查看本地是否有该区块，如果没有就验证区块和区块里的交易，通过之后就加入本地区块
				if !storage.HasBlockHash(hash) {
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

//todo:这里需要好好想想，怎样才算是真正的同步成功了呢？
func (node *Node) SyncMempool(wg *sync.WaitGroup) {
	//发送mempool消息，接收inv消息

	count := 0
	node.Peers.Range(func(key, value interface{}) bool {
		addr := key.(string)
		peer := value.(Peer)
		msg, err := p2p.NewMempoolMsg()
		if err != nil {
			panic(err) //如果连消息都会构造错误那只能panic了
		}
		if err = node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
			log.Error(err)
			return true //虽然出错了，但依然可以进行后面节点的处理
		}
		log.Infof("sync memory pool from [%s]", addr)

		count++
		if count >= MaxSyncPeerCount {
			return false
		}
		return true
	})
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
		peer.Alive <- true
	}

	return nil
}

func (node *Node) HandleTx(peer *Peer, payload []byte) error {
	_ = peer
	//todo:验证交易是否合法
	//如果交易合法，则查看本地是否有该交易，如果没有就加入本地交易池
	log.Debug("tx:", hex.EncodeToString(payload))
	txHash := common.Sha256AfterSha256(payload)

	//看来还挺难搞哦，这个可能刚刚被打包进区块了，所以交易池没有这个交易，如果就这么再添加到交易池的话，那就是很大的bug了。
	//要解决这种问题，看来只能把所有区块都同步下来并回，才能确定唯一性。
	node.txPool.Store(txHash, payload)
	return nil
}

func (node *Node) BroadcastNewBlock(blk *p2p.BlockPayload) {
	//把新的区块存入本地数据库，其实应该先存到内存，这样更快，这样可以以更快的速度广播新区块给相连的节点
	if err := storage.StoreSync(blk); err != nil { //这里必须使用同步存储，否则可能导致后面响应对方节点getdata的时候找不到这个block的bug
		log.Error(err)
		return
	}
	//删除本地mempool中与区块里重合的tx
	node.removeTxsInBlock(blk)

	//然后发inv消息给相连的节点询问他们是否需要这个最新的区块
	invpl := p2p.InvPayload{}
	invpl.Count = common.NewVarInt(uint64(1))
	blockhash := common.Sha256AfterSha256(blk.Header.Serialize())
	invpl.Inventory = append(invpl.Inventory, common.InvVector{common.MsgBlock, blockhash})
	invMsg, err := p2p.NewMsg("inv", invpl.Serialize())
	if err != nil {
		log.Error(err)
		return
	}

	node.Peers.Range(func(key, value interface{}) bool {
		peer := value.(Peer)
		if err = p2p.MustWrite(peer.Conn, invMsg.Serialize()); err != nil {
			log.Error(err)
		}
		return true
	})
}

func (node *Node) FetchTx(n uint32) []p2p.TxPayload {
	var txs []p2p.TxPayload
	var count = uint32(0)
	node.txPool.Range(func(key, value interface{}) bool {
		if count >= n { //放在for循环开头，解决n等于0的时候仍然获取交易的bug
			return true
		}
		tx := p2p.TxPayload{}
		err := tx.Parse(value.([]byte))
		if err != nil {
			log.Error(err)
			return false
		}
		txs = append(txs, tx)
		count++
		return true
	})

	if txs == nil {
		log.Info("tx mempool is empty")
		return nil
	}
	return txs
}

func BroadcastNewBlock(blk *p2p.BlockPayload) {
	defaultNode.BroadcastNewBlock(blk)
}

func FetchTx(n uint32) []p2p.TxPayload {
	return defaultNode.FetchTx(n)
}
