package block

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Header struct {
	BlockVersion   int32
	PreHash        [32]byte
	MerkleRootHash [32]byte
	Timestamp      uint32
	Bits           uint32
	Nonce          uint32
}

func (bh *Header) Len() int {
	return 80
}

func (bh *Header) Serialize() []byte {
	var data []byte
	var uint32Bytes [4]byte

	binary.LittleEndian.PutUint32(uint32Bytes[:], uint32(bh.BlockVersion))
	data = append(data, uint32Bytes[:]...)
	data = append(data, bh.PreHash[:]...)
	data = append(data, bh.MerkleRootHash[:]...)
	binary.LittleEndian.PutUint32(uint32Bytes[:], bh.Timestamp)
	data = append(data, uint32Bytes[:]...)
	binary.LittleEndian.PutUint32(uint32Bytes[:], bh.Bits)
	data = append(data, uint32Bytes[:]...)
	binary.LittleEndian.PutUint32(uint32Bytes[:], bh.Nonce)
	data = append(data, uint32Bytes[:]...)
	return data
}

func (bh *Header) String() string {
	var prettyJSON []byte
	var err error
	if prettyJSON, err = json.MarshalIndent(bh, "", "\t"); err != nil {
		return "parse block head struct failed"
	}

	return string(prettyJSON)
}

func (bh *Header) Parse(headerInRaw string) error {
	//参数校验
	if len(headerInRaw) < binary.Size(Header{})*2 { //每个字节用2个字符存储
		return errors.New("block header string is too short")
	}

	var (
		err        error
		tmp        []byte
		dataUint32 = make([]byte, 4)
	)

	if dataUint32, err = hex.DecodeString(headerInRaw[:8]); err != nil {
		return err
	}
	bh.BlockVersion = int32(binary.LittleEndian.Uint32(dataUint32))

	if tmp, err = hex.DecodeString(headerInRaw[8:72]); err != nil {
		return err
	}
	copy(bh.PreHash[:], tmp)

	if tmp, err = hex.DecodeString(headerInRaw[72:136]); err != nil {
		return err
	}
	copy(bh.MerkleRootHash[:], tmp)

	dataUint32, _ = hex.DecodeString(headerInRaw[136:144])
	bh.Timestamp = binary.LittleEndian.Uint32(dataUint32)

	dataUint32, _ = hex.DecodeString(headerInRaw[144:152])
	bh.Bits = binary.LittleEndian.Uint32(dataUint32)

	dataUint32, _ = hex.DecodeString(headerInRaw[152:])
	bh.Nonce = binary.LittleEndian.Uint32(dataUint32)

	return nil
}
