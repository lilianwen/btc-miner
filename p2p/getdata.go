package p2p

import "btcnetwork/common"

type GetdataPayload = InvPayload

func NewGetdataPayload(count uint64,invv []common.InvVector) *GetdataPayload {
	return &GetdataPayload{Count:common.NewVarInt(count), Inventory:invv}
}

func (node *Node)HandleGetdata(peer *Peer, payload []byte) error {
	//主要处理block和tx这两种数据，其他的暂时忽略
	ivp := InvPayload{}
	err := ivp.Parse(payload)
	if err != nil {
		return err
	}
	for i:=uint64(0); i<ivp.Count.Value; i++ {
		switch ivp.Inventory[i].Type {
		case common.MsgTx: {
			//todo:发送tx消息给对方

		}
		case common.MsgBlock: {
			//todo:发送block消息给对方

		}
		case common.MsgFilteredBlock:fallthrough
		case common.MsgCmpctBlock:fallthrough
		case common.MsgErr:fallthrough
		default:
			//什么都不做，忽略
		}
	}

	return nil
}
