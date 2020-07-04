package p2p

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

type NetAddr struct {
	Time     uint32 //Not present in version message,坑爹，在version消息里不显示
	Services uint64
	AddrIP   [16]byte
	Port     [2]byte
}

func (na *NetAddr) ToString() string {
	port := binary.BigEndian.Uint16(na.Port[:])
	return fmt.Sprintf("%d.%d.%d.%d:%d", na.AddrIP[12], na.AddrIP[13], na.AddrIP[14], na.AddrIP[15], port)
}

func (na *NetAddr) Len() int {
	return 26
}

func (na *NetAddr) Parse(data []byte) {
	na.Services = binary.LittleEndian.Uint64(data[:8])
	copy(na.AddrIP[:], data[8:24])
	copy(na.Port[:], data[24:26])
}

//返回大端序列
func IP2BytesUncheck(ip string) []byte {
	//参数校验（不校验了）
	var values []byte
	var strValues []string
	strValues = strings.Split(ip, ".")
	for _, elem := range strValues {
		n, _ := strconv.Atoi(elem) //不做错误校验
		values = append(values, byte(n))
	}
	return values
}

func NewNetAddr(timestamp uint32, service uint64, ip string, port uint16) NetAddr {
	var na = NetAddr{}
	na.Time = timestamp
	na.Services = 1
	na.AddrIP = [16]byte{00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0xFF, 0xFF, 00, 00, 00, 00}
	copy(na.AddrIP[12:], IP2BytesUncheck(ip))
	binary.BigEndian.PutUint16(na.Port[:], port)
	return na
}

func (na *NetAddr) Serialize() []byte {
	var data []byte
	var uint64Bytes [8]byte
	if na.Time != 0 { //这个是为了在version消息里不包含进去
		binary.LittleEndian.PutUint32(uint64Bytes[:4], na.Time)
		data = append(data, uint64Bytes[:4]...)
	}
	binary.LittleEndian.PutUint64(uint64Bytes[:], na.Services)
	data = append(data, uint64Bytes[:]...)
	data = append(data, na.AddrIP[:]...)
	data = append(data, na.Port[:]...)
	return data
}
