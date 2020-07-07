package p2p

import (
	"btcnetwork/common"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

type PreOutput struct {
	Hash  [32]byte
	Index uint32
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
	DataLen uint32
	Data    []byte
}

func (txw *TxWitness) Parse(data []byte) error {
	txw.DataLen = binary.LittleEndian.Uint32(data[:4])
	txw.Data = append(txw.Data, data[4:4+int(txw.DataLen)]...)
	return nil
}

func (txw *TxWitness) Len() int {
	return 4 + len(txw.Data)
}

type TxPayload struct {
	Version uint32
	//Flag uint16 //标志，如果存在就一定是0001
	TxinCount  common.VarInt
	Txins      []TxInput
	TxoutCount common.VarInt
	TxOuts     []TxOutput
	//WitnessCount //数量和txins的数量应该是相等的，坑爹的这和wiki上描述的结构不相同
	TxWitnesses []TxWitness
	Locktime    uint32
}

func (txp *TxPayload) Parse(data []byte) error {
	var isWitness = false
	var start = 0
	txp.Version = binary.LittleEndian.Uint32(data[:4])
	if data[4] == 0x00 && data[5] == 0x01 {
		//说明是隔离见证交易
		isWitness = true
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
	var out = TxOutput{}
	for i := 0; i < int(txp.TxoutCount.Value); i++ {
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

var TxPool map[[32]byte][]byte

func (node *Node) HandleTx(peer *Peer, payload []byte) error {
	_ = peer
	//todo:验证交易是否合法
	//如果交易合法，则查看本地是否有该交易，如果没有就加入本地交易池
	txHash := common.Sha256AfterSha256(payload)

	//看来还挺难搞哦，这个可能刚刚被打包进区块了，所以交易池没有这个交易，如果就这么再添加到交易池的话，那就是很大的bug了。
	//要解决这种问题，看来只能把所有区块都同步下来并回，才能确定唯一性。
	if _, ok := TxPool[txHash]; !ok { //todo:先暂时这么粗暴的做，这里肯定有bug
		TxPool[txHash] = payload
	}
	return nil
}

func init() {
	TxPool = make(map[[32]byte][]byte)
}
