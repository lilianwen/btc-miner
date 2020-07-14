package p2p

import "btcnetwork/common"

type GetdataPayload = InvPayload

func NewGetdataPayload(count uint64, invv []common.InvVector) *GetdataPayload {
	return &GetdataPayload{Count: common.NewVarInt(count), Inventory: invv}
}
