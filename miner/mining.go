package miner

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"btcnetwork/node"
	"btcnetwork/p2p"
	"btcnetwork/storage"
	"btcnetwork/transaction"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"sync"
	"time"
)

var (
	ErrNonceNotFound = errors.New("nonce not found")
)

func mineMonitor() {
	wg := sync.WaitGroup{}
deadloop:
	for {
		cmd := <-common.MinerCmd
		switch cmd {
		case common.MineOneTime:
			if minerConfig.state == StateAuto {
				log.Info("miner state is auto, not response to mine one block")
				continue
			}
			minerConfig.state = StateOneBlock
			mineOneBlock()
		case common.AutoMine:
			if minerConfig.state == StateAuto {
				log.Info("miner state is already auto state")
				continue
			}
			minerConfig.state = StateAuto
			go autoMine(&wg)
		case common.StopMine:
			minerConfig.state = StateStop
			break deadloop
		}
	}
	wg.Wait()
	minerStop <- true
}

func autoMine(wg *sync.WaitGroup) {
	for minerConfig.state == StateAuto {
		mineOneBlock()
	}
	wg.Done()
}

func requireTxs() []p2p.TxPayload {
	//获取交易
	txString := "01000000012b98ff080a16eafa8d696822c73a8ae547a5d41c08068aced0b529ac8353e7be000000006a473044022073393b92389248861bd68992303fc4a1ebab6ebc13c74b3b2692c140376fbec50220232983a24688fe5986694cd28879f2bf971d72385621f8987964004196f9541a0121033bee237c0e48aad2cc4411f7c51f424a94f063452ce32457d94be2d91e51f712ffffffff0200e1f505000000001976a91482d55da28ed20143e127b21c4aacc062424f46d388ac2010102401000000160014b7ad1b0d27ce120f9ce148f83a0734ef5f8f8b6700000000"

	buf, _ := hex.DecodeString(txString)
	tx := p2p.TxPayload{}
	_ = tx.Parse(buf)

	var txs []p2p.TxPayload
	txs = append(txs, tx)
	return txs
}

func createCoinbaseSigScript(height, extranonce uint64, banner string) []byte {
	var script []byte
	buf := CScriptNum(height)
	script = append(script, buf...)
	buf = CScriptNum(extranonce)
	script = append(script, buf...)
	return append(script, []byte(banner)...)
}

func CScriptNum(n uint64) []byte {
	var (
		i2b         []byte
		highestByte byte
		ret         []byte
	)
	for n > 0 {
		b := byte(n & 0xff)
		i2b = append(i2b, b)
		n = n >> 8
		highestByte = b
	}
	var extraByte []byte
	if highestByte&0x80 != 0 {
		extraByte = append(extraByte, byte(0x00))
	}
	i2b = append(i2b, extraByte...)
	ret = append(ret, byte(len(i2b)))
	ret = append(ret, i2b...)
	return ret
}

func createCoinbase(txs []p2p.TxPayload) *p2p.TxPayload {
	coinbase := p2p.TxPayload{}
	coinbase.Version = minerConfig.Version
	coinbase.Marker = nil
	coinbase.Flag = nil
	coinbase.TxinCount = common.NewVarInt(1)
	input := p2p.TxInput{}
	input.PreOut = p2p.NewCoinPreOutput()
	height := uint64(storage.LatestBlockHeight()) + uint64(1)
	//input.SigScript = []byte{0x02, 0x8a, 0x00, 0x08, 0x17, 0x21, 0xf2, 0xde, 0x15, 0x75, 0xae, 0x92, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48, 0x2f, 0x62, 0x74, 0x63, 0x64, 0x32} //先写死测试一下
	//input.SigScript[1] = byte(height)//临时权宜之计
	extraNonce, _ := rand.Prime(rand.Reader, 63) //这样得到的extraNonce肯定是正数
	input.SigScript = createCoinbaseSigScript(height, extraNonce.Uint64(), Banner)
	input.ScriptLen = common.NewVarInt(uint64(len(input.SigScript)))
	input.Sequence = 0xffffffff
	coinbase.Txins = append(coinbase.Txins, input)
	coinbase.TxoutCount = common.NewVarInt(1)
	output := p2p.TxOutput{}
	output.Value = minerConfig.Reward + storage.Fee(txs) //由2部分构成，一部分是系统奖励，另一部分是交易手续费
	output.PkScript = transaction.NewP2PKHScipt(minerConfig.MinerPubKeyHash[:])
	output.PkScriptLen = common.NewVarInt(uint64(len(output.PkScript)))
	coinbase.TxOuts = append(coinbase.TxOuts, output)
	coinbase.WitnessCount = nil
	coinbase.TxWitnesses = nil
	coinbase.Locktime = 0
	return &coinbase
}

func mineOneBlock() {
	var (
		txids    []string
		txid     [32]byte
		txs      = node.FetchTx(1) //暂时只提取一个交易进行打包
		coinbase = createCoinbase(txs)
	)

	txid = coinbase.Txid()
	txids = append(txids, hex.EncodeToString(common.ReverseBytes(txid[:])))
	for i := 0; i < len(txs); i++ {
		txid = txs[i].Txid()
		txids = append(txids, hex.EncodeToString(common.ReverseBytes(txid[:])))
	}

	root, _ := block.ConstructMerkleRoot(txids)
	buf, err := hex.DecodeString(root.Value)
	if err != nil {
		log.Errorf("merkle root hash decode error:%s", root.Value)
	}
	common.ReverseBytes(buf)

	header := block.Header{}
	header.BlockVersion = 0x05
	header.PreHash = storage.LatestBlockHash()
	copy(header.MerkleRootHash[:], buf) //注意：这里的值要不要逆序？
	header.Timestamp = uint32(time.Now().Unix())
	header.Bits = minerConfig.Bits
	nonce, err := searchNonce(&header, 0)
	if err != nil {
		//前期挖矿失败，就不挖了，todo:后期考虑调整交易顺序，使用扩展nonce等策略
		log.Error("not found avaliable nonce")
	} else {
		header.Nonce = nonce
		//组建区块，广播给其他节点
		blk := p2p.BlockPayload{}
		blk.Header = header
		blk.TxnCount = common.NewVarInt(uint64(1 + len(txs)))
		blk.Txns = append(blk.Txns, *coinbase)
		for i := 0; i < len(txs); i++ {
			blk.Txns = append(blk.Txns, txs[i])
		}
		node.BroadcastNewBlock(&blk)

		//刷新minerConfig,每2016个区块调整难度值，可能需要调整难度值
		log.Infof("generate block: nonce=%d", nonce)
	}
}

func searchNonce(header *block.Header, startNonce uint32) (uint32, error) {
	wantTarget := block.Bits2Target(header.Bits)
	log.Debug("start mining with target ", hex.EncodeToString(wantTarget.Bytes()))
	buf := header.Serialize()
	for i := startNonce; true; i++ {
		var i2b4 [4]byte
		binary.LittleEndian.PutUint32(i2b4[:], i)
		copy(buf[76:], i2b4[:])
		blockHash := common.Sha256AfterSha256(buf)

		common.ReverseBytes(blockHash[:]) //注意：这里一定要反转一下顺序,因为big.Int是大端存储
		gotTarget := new(big.Int).SetBytes(blockHash[:])
		if wantTarget.Cmp(gotTarget) >= 0 { //bingo 挖到区块了
			return i, nil
		}
		if i == math.MaxUint32 {
			break
		}
		if i%10000000 == 0 {
			log.Info("try nonce: ", i)
		}
		if minerConfig.state == StateStop { //需要立刻停止挖矿
			return 0, ErrNonceNotFound
		}
	}
	return 0, ErrNonceNotFound
}

func Integer2bytes(i int32) []byte {
	var buf []byte
	for i > 0 {
		n := i & 0xff
		buf = append(buf, byte(n))
		i = i >> 8
	}
	return buf
}
