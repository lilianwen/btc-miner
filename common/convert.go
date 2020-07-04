package common

import (
	"encoding/binary"
)

// 把单个uint32类型数据变成大端字节序
func Uint32ToBytes(n uint32) []byte {
	var uint32Bytes [4]byte
	binary.BigEndian.PutUint32(uint32Bytes[:], n)
	return uint32Bytes[:]
}

func ReverseBytes(data []byte) {
	var length = len(data)
	for i := 0; i < length/2; i++ {
		data[i], data[length-1-i] = data[length-1-i], data[i]
	}
}

func Byte2String(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}
