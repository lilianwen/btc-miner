package p2p

import (
	"btcnetwork/common"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type VersionPayload struct {
	Version     int32
	Services    uint64
	Timestamp   int64
	AddrRecv    NetAddr
	AddrFrom    NetAddr
	Nonce       uint64
	UserAgent   VarStr //存放子版本号
	StartHeight int32
	Relay       bool //when version >= 70001 require this
}

func (vp *VersionPayload) Parse(data []byte) {
	//参数校验，我靠，这怎么校验？
	vp.Version = int32(binary.LittleEndian.Uint32(data[:4]))
	vp.Services = binary.LittleEndian.Uint64(data[4:12])
	vp.Timestamp = int64(binary.LittleEndian.Uint64(data[12:20]))
	vp.AddrRecv.Parse(data[20:46])
	vp.AddrFrom.Parse(data[46:72])
	vp.Nonce = binary.LittleEndian.Uint64(data[72:80])
	vp.UserAgent.Parse(data[80:])
	vp.StartHeight = int32(binary.LittleEndian.Uint32(data[80+vp.UserAgent.Len():]))
	vp.Relay = data[80+vp.UserAgent.Len()+4] == byte(1)
}

func (vp *VersionPayload) Serialize() []byte {
	var data []byte
	var uint32Bytes [4]byte
	var uint64Bytes [8]byte
	var buf []byte
	binary.LittleEndian.PutUint32(uint32Bytes[:], uint32(vp.Version))
	data = append(data, uint32Bytes[:]...)
	binary.LittleEndian.PutUint64(uint64Bytes[:], vp.Services)
	data = append(data, uint64Bytes[:]...)
	binary.LittleEndian.PutUint64(uint64Bytes[:], uint64(vp.Timestamp))
	data = append(data, uint64Bytes[:]...)
	buf = vp.AddrRecv.Serialize()
	data = append(data, buf...)
	buf = vp.AddrFrom.Serialize()
	data = append(data, buf...)
	binary.LittleEndian.PutUint64(uint64Bytes[:], vp.Nonce)
	data = append(data, uint64Bytes[:]...)
	buf = vp.UserAgent.Serialize()
	data = append(data, buf...)
	binary.LittleEndian.PutUint32(uint32Bytes[:], uint32(vp.StartHeight))
	data = append(data, uint32Bytes[:]...)
	if vp.Relay {
		data = append(data, byte(1))
	} else {
		data = append(data, byte(0))
	}
	return data
}

//是uncheck还是unchecked？
func NewVersionPayloadUncheck(ipAddr string, port uint16) *VersionPayload {
	//参数校验
	var vp = VersionPayload{}
	vp.Version = 70002
	vp.Services = 1
	vp.Timestamp = time.Now().Unix()
	vp.AddrFrom = NewNetAddr(0, vp.Services, "127.0.0.1", 8333)
	vp.AddrRecv = NewNetAddr(0, vp.Services, ipAddr, port)
	vp.Nonce = rand.Uint64()

	vp.UserAgent = NewSubVersion()
	vp.StartHeight = int32(0)
	vp.Relay = true
	return &vp
}

func NewVerMsg(address string) (*Msg, error) {
	addr := strings.Split(address, ":")
	ip := addr[0]
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		return nil, err
	}
	payload := NewVersionPayloadUncheck(ip, uint16(port))
	return NewMsg("version", payload.Serialize())
}

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(common.LogLevel)
}
