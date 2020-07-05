package p2p

import (
	"btcnetwork/common"
)

type InvPayload struct {
	Count common.VarInt
	Inventory []common.InvVector
}

func (invp *InvPayload)Serialize() []byte {
	var buf []byte
	buf = append(buf, invp.Count.Data...)
	for i:= range invp.Inventory {
		buf = append(buf, invp.Inventory[i].Serialize()...)
	}
	return buf
}

func (invp *InvPayload)Parse(data []byte) error {
	return nil
}

