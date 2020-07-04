package p2p

import (
	"btcnetwork/common"
	"encoding/hex"
	"io/ioutil"
	"net"
	"net/http"
)

type VarStr struct {
	Length common.VarInt
	Data   string
}

func (vs *VarStr) Parse(data []byte) {
	vs.Length.Parse(hex.EncodeToString(data))
	var startIndex = vs.Length.Len()
	var endIndex = startIndex + int(vs.Length.Value)
	vs.Data = string(data[startIndex:endIndex])
}

func (vs *VarStr) Len() int {
	return vs.Length.Len() + len(vs.Data)
}

func NewSubVersion() VarStr {
	var sv = VarStr{}
	sv.Data = "/Satoshi:0.8.6/"
	sv.Length = common.NewVarInt(uint64(len(sv.Data)))
	return sv
}

func (vs *VarStr) Serialize() []byte {
	var data []byte
	data = append(data, vs.Length.Data...)
	data = append(data, vs.Data...)
	return data
}

func MustWrite(conn net.Conn, buf []byte) error {
	var (
		n         = 0
		sum       = 0
		err error = nil
	)

	for {
		if n, err = conn.Write(buf[sum:]); err != nil {
			return err
		}
		sum += n
		if sum == len(buf) {
			break
		}
	}
	return nil
}

func GetExternalIP() string {
	//为提高速度，直接硬编码
	//return "112.97.56.108"

	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)

}
