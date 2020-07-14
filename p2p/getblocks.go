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
	var buf []byte
	var uint32Buf [4]byte

	binary.LittleEndian.PutUint32(uint32Buf[:], gp.Version)

	buf = append(buf, uint32Buf[:]...)
	buf = append(buf, gp.HashCount.Data...)
	buf = append(buf, gp.HashStart[:]...)
	buf = append(buf, gp.HashStop[:]...)

	return buf
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
