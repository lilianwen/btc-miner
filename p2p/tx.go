package p2p

import "btcnetwork/common"

//type TxPayload struct {
//	Version uint32
//	Flag uint16 //标志，如果存在就一定是0001
//	TxinCount common.VarInt
//	Txins []TxInput
//	TxoutCount common.VarInt
//	TxOuts []TxOutput
//	TxWitnesses []TxWitness
//	Locktime uint32
//}

var TxPool map[[32]byte][]byte

func (node *Node) HandleTx(peer *Peer, payload []byte) error {
	_ = peer
	//todo:验证交易是否合法
	//如果交易合法，则查看本地是否有该交易，如果没有就加入本地交易池
	txHash := common.Sha256AfterSha256(payload)

	//看来还挺难搞哦，这个可能刚刚被打包进区块了，所以交易池没有这个交易，如果就这么再添加到交易池的话，那就是很大的bug了。
	//要解决这种问题，看来只能把所有区块都同步下来并回，才能确定唯一性。
	if _, ok := TxPool[txHash]; !ok { //todo:先暂时这么粗暴的做，这里肯定有bug
		TxPool[txHash] = payload
	}
	return nil
}

func init() {
	TxPool = make(map[[32]byte][]byte)
}
