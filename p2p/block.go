package p2p

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"encoding/hex"
)

type BlockPayload struct {
	block.Header
	TxnCount common.VarInt
	Txns     []TxPayload
}

func (bp *BlockPayload) Parse(data []byte) error {
	log.Info(hex.EncodeToString(data))
	err := bp.Header.Parse(hex.EncodeToString(data))
	if err != nil {
		return err
	}
	start := bp.Header.Len()
	if err = bp.TxnCount.Parse(hex.EncodeToString(data[start:])); err != nil {
		return err
	}
	start += bp.TxnCount.Len()
	for i := uint64(0); i < bp.TxnCount.Value; i++ {
		tx := TxPayload{}
		if err = tx.Parse(data[start:]); err != nil {
			return err
		}
		bp.Txns = append(bp.Txns, tx)
		start += tx.Len()
	}
	return nil
}
