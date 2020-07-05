package common

import (
	"encoding/binary"
	"errors"
)

type ObjectType uint32
const (
	MsgErr ObjectType = iota
	MsgTx
	MsgBlock
	MsgFilteredBlock
	MsgCmpctBlock
)

type InvVector struct {
	Type ObjectType
	Hash [32]byte
}

func NewInvVector(t ObjectType, hash []byte) *InvVector{
	iv := InvVector{}
	iv.Type = t
	copy(iv.Hash[:], hash[:12])
	return &iv
}

func (iv *InvVector)Serialize() []byte {
	var buf []byte
	var uint32Bytes [4]byte
	binary.LittleEndian.PutUint32(uint32Bytes[:], uint32(iv.Type))
	buf = append(buf, uint32Bytes[:]...)
	buf = append(buf, iv.Hash[:]...)
	return buf
}

func (iv *InvVector)Parse(data []byte) error {
	if len(data) < iv.Len() {
		return errors.New("data is too short")
	}
	iv.Type = ObjectType(binary.LittleEndian.Uint32(data[:4]))
	copy(iv.Hash[:], data[4:36])
	return nil
}

func (iv *InvVector)Len() int {
	return 36
}
