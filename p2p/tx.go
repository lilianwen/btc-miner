package p2p

import (
	"btcnetwork/common"
	"reflect"

	//"btcnetwork/storage"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

type PreOutput struct {
	Hash  [32]byte
	Index uint32
}

func NewCoinPreOutput() PreOutput {
	preOutput := PreOutput{}
	preOutput.Index = 0xffffffff
	return preOutput
}

func (po *PreOutput) Serialize() []byte {
	var i2b4 [4]byte
	var ret []byte

	ret = append(ret, po.Hash[:]...)
	binary.LittleEndian.PutUint32(i2b4[:], po.Index)
	ret = append(ret, i2b4[:]...)
	return ret
}

func (po *PreOutput) Parse(data []byte) error {
	if len(data) < 32 {
		return errors.New("data is too short for pre output hash")
	}
	copy(po.Hash[:], data[:32])
	po.Index = binary.LittleEndian.Uint32(data[32:36])
	return nil
}
func (po *PreOutput) Len() int {
	return 36
}

type TxInput struct {
	PreOut    PreOutput
	ScriptLen common.VarInt
	SigScript []byte
	Sequence  uint32
}

func (txin *TxInput) Serialize() []byte {
	var ret []byte
	var buf = txin.PreOut.Serialize()
	var i2b4 [4]byte
	ret = append(ret, buf...)
	ret = append(ret, txin.ScriptLen.Data...)
	ret = append(ret, txin.SigScript...)
	binary.LittleEndian.PutUint32(i2b4[:], txin.Sequence)
	ret = append(ret, i2b4[:]...)

	return ret
}

func (txin *TxInput) Parse(data []byte) error {
	err := txin.PreOut.Parse(data)
	if err != nil {
		return err
	}
	start := txin.PreOut.Len()
	if err = txin.ScriptLen.Parse(hex.EncodeToString(data[start:])); err != nil {
		return err
	}

	start += txin.ScriptLen.Len()
	txin.SigScript = append(txin.SigScript, data[start:start+int(txin.ScriptLen.Value)]...)
	start += int(txin.ScriptLen.Value)
	txin.Sequence = binary.LittleEndian.Uint32(data[start : start+4])
	return nil
}
func (txin *TxInput) Len() int {
	return txin.PreOut.Len() + txin.ScriptLen.Len() + len(txin.SigScript) + 4
}

type TxOutput struct {
	Value       uint64
	PkScriptLen common.VarInt
	PkScript    []byte
}

func (txout *TxOutput) Serialize() []byte {
	var ret []byte
	var i2b8 [8]byte
	binary.LittleEndian.PutUint64(i2b8[:], txout.Value)
	ret = append(ret, i2b8[:]...)
	ret = append(ret, txout.PkScriptLen.Data...)
	ret = append(ret, txout.PkScript...)
	return ret
}

func (txout *TxOutput) Parse(data []byte) error {
	txout.Value = binary.LittleEndian.Uint64(data[:8])
	err := txout.PkScriptLen.Parse(hex.EncodeToString(data[8:]))
	if err != nil {
		return err
	}
	start := 8 + txout.PkScriptLen.Len()
	txout.PkScript = append(txout.PkScript, data[start:start+int(txout.PkScriptLen.Value)]...)
	return nil
}

func (txout *TxOutput) Len() int {
	return 8 + txout.PkScriptLen.Len() + len(txout.PkScript)
}

type TxWitness struct {
	DataLen common.VarInt
	Data    []byte
}

func (txw *TxWitness) Serialize() []byte {
	var ret []byte
	ret = append(ret, txw.DataLen.Data...)
	ret = append(ret, txw.Data...)
	return ret
}

func (txw *TxWitness) Parse(data []byte) error {
	err := txw.DataLen.Parse(hex.EncodeToString(data))
	if err != nil {
		return err
	}
	start := txw.DataLen.Len()
	txw.Data = append(txw.Data, data[start:start+int(txw.DataLen.Value)]...)
	return nil
}

func (txw *TxWitness) Len() int {
	return txw.DataLen.Len() + len(txw.Data)
}

type TxPayload struct {
	Version      uint32
	Marker       []byte //隔离见证标志，如果存在就一定是00
	Flag         []byte //隔离见证标志，如果存在就一定是01
	TxinCount    common.VarInt
	Txins        []TxInput
	TxoutCount   common.VarInt
	TxOuts       []TxOutput
	WitnessCount []common.VarInt //数量和txins的数量应该是相等的，坑爹的这和wiki上描述的结构不相同
	TxWitnesses  []TxWitness
	Locktime     uint32
}

func (txp *TxPayload) CoinbaseTx() bool {
	var zeroHash [32]byte
	return reflect.DeepEqual(txp.Txins[0].PreOut.Hash[:], zeroHash[:])
}

func (txp *TxPayload) TxHash() [32]byte {
	var i2b4 [4]byte
	var ret []byte
	binary.LittleEndian.PutUint32(i2b4[:], txp.Version)
	ret = append(ret, i2b4[:]...)
	if len(txp.Marker) != 0 {
		ret = append(ret, txp.Marker...)
	}
	if len(txp.Flag) != 0 {
		ret = append(ret, txp.Flag...)
	}

	ret = append(ret, txp.TxinCount.Data...)
	for i := uint64(0); i < txp.TxinCount.Value; i++ {
		in := txp.Txins[i].Serialize()
		ret = append(ret, in...)
	}
	ret = append(ret, txp.TxoutCount.Data...)
	for i := uint64(0); i < txp.TxoutCount.Value; i++ {
		out := txp.TxOuts[i].Serialize()
		ret = append(ret, out...)
	}
	if len(txp.WitnessCount) != 0 {
		ret = append(ret, txp.WitnessCount[0].Data...)
		for i := uint64(0); i < txp.WitnessCount[0].Value; i++ {
			witness := txp.TxWitnesses[i].Serialize()
			ret = append(ret, witness...)
		}
	}

	binary.LittleEndian.PutUint32(i2b4[:], txp.Locktime)
	ret = append(ret, i2b4[:]...)
	log.Debug("tx serialized:", hex.EncodeToString(ret))
	return common.Sha256AfterSha256(ret)
}

func (txp *TxPayload) Txid() [32]byte {
	var i2b4 [4]byte
	var ret []byte
	binary.LittleEndian.PutUint32(i2b4[:], txp.Version)
	ret = append(ret, i2b4[:]...)
	ret = append(ret, txp.TxinCount.Data...)
	for i := uint64(0); i < txp.TxinCount.Value; i++ {
		in := txp.Txins[i].Serialize()
		ret = append(ret, in...)
	}
	ret = append(ret, txp.TxoutCount.Data...)
	for i := uint64(0); i < txp.TxoutCount.Value; i++ {
		out := txp.TxOuts[i].Serialize()
		ret = append(ret, out...)
	}

	binary.LittleEndian.PutUint32(i2b4[:], txp.Locktime)
	ret = append(ret, i2b4[:]...)
	return common.Sha256AfterSha256(ret)
}

func (txp *TxPayload) Len() int {
	var length = 4 + len(txp.Marker) + len(txp.Flag) + txp.TxinCount.Len()
	for i := uint64(0); i < txp.TxinCount.Value; i++ {
		length += txp.Txins[i].Len()
	}
	length += txp.TxoutCount.Len()
	for i := uint64(0); i < txp.TxoutCount.Value; i++ {
		length += txp.TxOuts[i].Len()
	}
	if len(txp.WitnessCount) != 0 {
		length += txp.WitnessCount[0].Len()
		for i := uint64(0); i < txp.WitnessCount[0].Value; i++ {
			length += txp.TxWitnesses[i].Len()
		}
	}
	length += 4
	return length
}

func (txp *TxPayload) Parse(data []byte) error {
	log.Debug("txpayload size:", len(data))
	log.Debug("tx data:", hex.EncodeToString(data))
	txhash := common.Sha256AfterSha256(data)
	log.Debug("tx hash:", hex.EncodeToString(txhash[:]))
	var isWitness = false
	var start = 0
	txp.Version = binary.LittleEndian.Uint32(data[:4])
	if data[4] == 0x00 && data[5] == 0x01 {
		//说明是隔离见证交易
		isWitness = true
		txp.Marker = append(txp.Marker, data[4])
		txp.Flag = append(txp.Flag, data[5])
		start = 6
	} else {
		start = 4
	}
	err := txp.TxinCount.Parse(hex.EncodeToString(data[start:]))
	if err != nil {
		return err
	}
	start += txp.TxinCount.Len()
	var in = TxInput{}
	for i := 0; i < int(txp.TxinCount.Value); i++ {
		if err = in.Parse(data[start:]); err != nil {
			return err
		}
		start += in.Len()
		txp.Txins = append(txp.Txins, in)
	}
	if err = txp.TxoutCount.Parse(hex.EncodeToString(data[start:])); err != nil {
		return err
	}
	start += txp.TxoutCount.Len()
	for i := 0; i < int(txp.TxoutCount.Value); i++ {
		var out = TxOutput{}
		if err = out.Parse(data[start:]); err != nil {
			return err
		}
		start += out.Len()
		txp.TxOuts = append(txp.TxOuts, out)
	}

	if isWitness {
		var witnessCount common.VarInt
		if err = witnessCount.Parse(hex.EncodeToString(data[start:])); err != nil {
			return err
		}
		if witnessCount.Value != txp.TxinCount.Value {
			return errors.New("witness count != tx input count")
		}
		txp.WitnessCount = append(txp.WitnessCount, witnessCount)
		start += witnessCount.Len()
		var witness TxWitness
		for i := 0; i < int(witnessCount.Value); i++ {
			if err = witness.Parse(data[start:]); err != nil {
				return err
			}
			start += witness.Len()
			txp.TxWitnesses = append(txp.TxWitnesses, witness)
		}
	}
	txp.Locktime = binary.LittleEndian.Uint32(data[start : start+4])
	return nil
}

func (txp *TxPayload) Serialize() []byte {
	var (
		ret  []byte
		i2b4 [4]byte
	)
	binary.LittleEndian.PutUint32(i2b4[:], txp.Version)
	ret = append(ret, i2b4[:]...)
	ret = append(ret, txp.Marker...)
	ret = append(ret, txp.Flag...)
	ret = append(ret, txp.TxinCount.Data...)
	for i := uint64(0); i < txp.TxinCount.Value; i++ {
		txinBytes := txp.Txins[i].Serialize()
		ret = append(ret, txinBytes...)
	}
	ret = append(ret, txp.TxoutCount.Data...)
	for i := uint64(0); i < txp.TxoutCount.Value; i++ {
		txoutBytes := txp.TxOuts[i].Serialize()
		ret = append(ret, txoutBytes...)
	}
	if len(txp.WitnessCount) != 0 {
		ret = append(ret, txp.WitnessCount[0].Data...)
		for i := uint64(0); i < txp.WitnessCount[0].Value; i++ {
			witnessBytes := txp.TxWitnesses[i].Serialize()
			ret = append(ret, witnessBytes...)
		}
	}
	binary.LittleEndian.PutUint32(i2b4[:], txp.Locktime)
	ret = append(ret, i2b4[:]...)

	return ret
}
