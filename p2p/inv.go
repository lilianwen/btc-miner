package p2p

import (
	"btcnetwork/common"
	"encoding/hex"
)

type InvPayload struct {
	Count     common.VarInt
	Inventory []common.InvVector
}

//func NewInvPayload(count uint64, invv []common.InvVector) *InvPayload {
////	return &InvPayload{Count: common.NewVarInt(count), Inventory: invv}
////}

func (invp *InvPayload) Serialize() []byte {
	var buf []byte
	buf = append(buf, invp.Count.Data...)
	for i := range invp.Inventory {
		buf = append(buf, invp.Inventory[i].Serialize()...)
	}
	return buf
}

func (invp *InvPayload) Parse(data []byte) error {
	err := invp.Count.Parse(hex.EncodeToString(data[:]))
	if err != nil {
		return err
	}
	log.Println("inv count:", invp.Count.Value)

	invp.Inventory = make([]common.InvVector, invp.Count.Value) //预先分配，提高性能
	start := invp.Count.Len()
	invvect := common.InvVector{}
	for i := uint64(0); i < invp.Count.Value; i++ {
		if err := invvect.Parse(data[start:]); err != nil {
			return err
		}
		log.Printf("inv[%d] type:%s", i, common.ObjectType2String(invvect.Type))
		log.Printf("inv[%d] hash:%s", i, hex.EncodeToString(invvect.Hash[:]))

		invp.Inventory[i] = invvect
		start += invvect.Len()
	}
	return nil
}
