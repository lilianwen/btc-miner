package p2p

import (
	"btcnetwork/common"
	"encoding/hex"
	"log"
)

type InvPayload struct {
	Count     common.VarInt
	Inventory []common.InvVector
}

func NewInvPayload(count uint64, invv []common.InvVector) *InvPayload {
	return &InvPayload{Count: common.NewVarInt(count), Inventory: invv}
}

func (invp *InvPayload) Serialize() []byte {
	var buf []byte
	buf = append(buf, invp.Count.Data...)
	for i := range invp.Inventory {
		buf = append(buf, invp.Inventory[i].Serialize()...)
	}
	return buf
}

func (invp *InvPayload) Parse(data []byte) error {
	invp.Count.Parse(hex.EncodeToString(data[:]))
	log.Println("inv count:", invp.Count.Value)

	invp.Inventory = make([]common.InvVector, invp.Count.Value) //预先分配，提高性能
	start := invp.Count.Len()
	invvect := common.InvVector{}
	for i := uint64(0); i < invp.Count.Value; i++ {
		if err := invvect.Parse(data[start:]); err != nil {
			return err
		}
		log.Printf("inv[%d] type:%s\n", i, common.ObjectType2String(invvect.Type))
		log.Printf("inv[%d] hash:%s\n", i, hex.EncodeToString(invvect.Hash[:]))

		invp.Inventory[i] = invvect
		start += invvect.Len()
	}
	return nil
}

func (node *Node) HandleInv(peer *Peer, payload []byte) error {
	//主要处理block和tx这两种数据，其他的暂时忽略
	ivp := InvPayload{}
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
				if _, ok := TxPool[hash]; !ok {
					//本地没有这个tx,发送getdata消息给节点获取交易数据
					//如果不想知道交易详情，其实是可以不去get交易数据的
					//toGetData = true
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
		msg, err := NewMsg("getdata", NewGetdataPayload(count, invVect).Serialize())
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
	return false
}
