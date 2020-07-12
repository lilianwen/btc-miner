package p2p

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"encoding/hex"
)

type BlockPayload struct {
	block.Header
	TxnCount common.VarInt
	Txns     []TxPayload
}

func (bp *BlockPayload) Parse(data []byte) error {
	log.Info(hex.EncodeToString(data))
	err := bp.Header.Parse(hex.EncodeToString(data))
	if err != nil {
		return err
	}
	start := bp.Header.Len()
	if err = bp.TxnCount.Parse(hex.EncodeToString(data[start:])); err != nil {
		return err
	}
	start += bp.TxnCount.Len()
	for i := uint64(0); i < bp.TxnCount.Value; i++ {
		tx := TxPayload{}
		if err = tx.Parse(data[start:]); err != nil {
			return err
		}
		bp.Txns = append(bp.Txns, tx)
		start += tx.Len()
	}
	return nil
}

func (node *Node) HandleBlock(peer *Peer, payload []byte) error {
	log.Infof("receive block data from [%s]", peer.Addr)

	recvBlock := BlockPayload{}
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

		strHash := hex.EncodeToString(common.ReverseBytes(txHash[:]))
		log.Infof("tx[%d].txhash=%s, size=%d", i, hex.EncodeToString(txHash[:]), recvBlock.Txns[i].Len())
		txid := recvBlock.Txns[i].Txid()
		log.Infof("tx[%d].txid=%s", i, hex.EncodeToString(txid[:]))
		hashes = append(hashes, strHash)

	}
	//尝试构建默克尔树
	root, err := block.ConstructMerkleRoot(hashes)
	wantMerkleRootHash := hex.EncodeToString(recvBlock.MerkleRootHash[:])
	if root.Value != wantMerkleRootHash {
		log.Error("calculate merkle root hash not equal to block header merkle root hash")
		log.Errorf("get:%s, want:%s", root.Value, wantMerkleRootHash)
	}
	return nil
}
