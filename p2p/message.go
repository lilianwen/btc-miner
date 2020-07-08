package p2p

import (
	"btcnetwork/common"
	"encoding/binary"
)

type Msg struct {
	Header  MsgHeader
	Payload []byte
}

func NewMsg(cmd string, data []byte) (*Msg, error) {
	//参数校验
	msg := Msg{}
	var err error
	if msg.Header, err = NewMsgHeader(cmd); err != nil {
		return nil, err
	}

	msg.Payload = make([]byte, len(data))
	copy(msg.Payload, data)
	hash256 := common.Sha256AfterSha256(msg.Payload)
	msg.Header.LenOfPayload = uint32(len(msg.Payload))
	msg.Header.Checksum = binary.LittleEndian.Uint32(hash256[:4])
	return &msg, nil
}

func (msg *Msg) Serialize() []byte {
	var data []byte
	var buf [4]byte

	binary.LittleEndian.PutUint32(buf[:], msg.Header.Magic)
	data = append(data, buf[:]...)
	data = append(data, msg.Header.Command[:]...)
	binary.LittleEndian.PutUint32(buf[:], msg.Header.LenOfPayload)
	data = append(data, buf[:]...)
	binary.LittleEndian.PutUint32(buf[:], msg.Header.Checksum)
	data = append(data, buf[:]...)
	data = append(data, msg.Payload...)
	return data
}

func (msg *Msg) Parse(data []byte) {
	msg.Header.Parse(data)
	msg.Payload = append(msg.Payload, data[24:]...)
}
