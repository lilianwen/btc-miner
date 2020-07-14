package p2p

import (
	"btcnetwork/common"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
)

type GetblocksPayload struct {
	Version   uint32
	HashCount common.VarInt
	HashStart [32]byte
	HashStop  [32]byte
}

func (gp *GetblocksPayload) Serialize() []byte {
	var ret []byte
	var i2b4 [4]byte

	binary.LittleEndian.PutUint32(i2b4[:], gp.Version)

	ret = append(ret, i2b4[:]...)
	ret = append(ret, gp.HashCount.Data...)
	ret = append(ret, gp.HashStart[:]...)
	ret = append(ret, gp.HashStop[:]...)

	return ret
}

func (gp *GetblocksPayload) Parse(data []byte) error {
	if len(data) < gp.Len() {
		return errors.New("data is too short for getblocks payload")
	}
	binary.LittleEndian.Uint32(data[:4])
	if err := gp.HashCount.Parse(hex.EncodeToString(data[4:])); err != nil {
		return err
	}
	start := 4 + gp.HashCount.Len()
	copy(gp.HashStart[:], data[start:start+32])
	start += 32
	copy(gp.HashStop[:], data[start:start+32])
	return nil
}

func (gp *GetblocksPayload) Len() int {
	return 68 + gp.HashCount.Len()
}
