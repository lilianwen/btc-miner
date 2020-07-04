package p2p

import (
	"encoding/binary"
	"errors"
)

const (
	MsgHeaderLen = 24 //消息头部的长度
)

type MsgHeader struct {
	Magic        uint32
	Command      [12]byte
	LenOfPayload uint32
	Checksum     uint32
}

func NewMsgHeader(cmd string) (MsgHeader, error) {
	if len(cmd) > 12 {
		return MsgHeader{}, errors.New("message command is too long")
	}
	var header = MsgHeader{}
	//header.Magic = uint32(0xD9B4BEF9)//mainnet
	header.Magic = uint32(0x12141c16)
	command := []byte(cmd)
	copy(header.Command[:], command)
	header.LenOfPayload = uint32(0)
	header.Checksum = uint32(0)
	return header, nil
}

func (header *MsgHeader) Parse(data []byte) {
	header.Magic = binary.LittleEndian.Uint32(data[:4])
	copy(header.Command[:], data[4:16])
	header.LenOfPayload = binary.LittleEndian.Uint32(data[16:20])
	header.Checksum = binary.LittleEndian.Uint32(data[20:24])
}
