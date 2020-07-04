package common

import (
	"encoding/binary"
	"encoding/hex"
)

type VarInt struct {
	Data  []byte
	Value uint64
}

func (vi *VarInt) Len() int {
	return len(vi.Data)
}

func (vi *VarInt) Parse(dataInRaw string) error {
	var (
		err error
		tmp []byte
	)
	if vi.Data, err = hex.DecodeString(dataInRaw[:2]); err != nil {
		return err
	}
	//清空tmp
	tmp = tmp[0:0]
	switch {
	case vi.Data[0] <= 0xfc:
		vi.Value = uint64(vi.Data[0])
	case vi.Data[0] == 0xfd:
		//后面两个字节
		if tmp, err = hex.DecodeString(dataInRaw[2:6]); err != nil {
			return err
		}
		vi.Value = uint64(binary.LittleEndian.Uint16(tmp))
	case vi.Data[0] == 0xfe:
		//后面四个字节
		if tmp, err = hex.DecodeString(dataInRaw[2:10]); err != nil {
			return err
		}
		vi.Value = uint64(binary.LittleEndian.Uint32(tmp))
	case vi.Data[0] == 0xff:
		//后面八个字节
		if tmp, err = hex.DecodeString(dataInRaw[2:18]); err != nil {
			return err
		}
		vi.Value = uint64(binary.LittleEndian.Uint64(tmp))
	}
	vi.Data = append(vi.Data, tmp...)
	return nil
}

func NewVarInt(value uint64) VarInt {
	var data []byte

	switch {
	case value <= 0xfc:
		data = append(data, byte(value))
	case value <= 0xffff:
		data = make([]byte, 3)
		data[0] = 0xfd
		binary.LittleEndian.PutUint16(data[1:], uint16(value))
	case value <= 0xffffffff:
		data = make([]byte, 5)
		data[0] = 0xfe
		binary.LittleEndian.PutUint32(data[1:], uint32(value))
	default:
		data = make([]byte, 9)
		data[0] = 0xff
		binary.LittleEndian.PutUint64(data[1:], value)
	}

	return VarInt{Value: value, Data: data}
}
