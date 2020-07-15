package storage

import (
	"btcnetwork/common"
	"btcnetwork/p2p"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"reflect"
)

const (
	UtxoTxChanSize = 2000
)

type utxoMgr struct {
	stop chan bool
	done chan bool
	tx   chan p2p.TxPayload

	dbUtxo *leveldb.DB
}

var defaultUtxoMgr *utxoMgr

func newUtxoMgr(cfg *common.Config) *utxoMgr {
	mgr := utxoMgr{}
	mgr.stop = make(chan bool, 1)
	mgr.done = make(chan bool, 1)
	mgr.tx = make(chan p2p.TxPayload, UtxoTxChanSize)
	var err error
	mgr.dbUtxo, err = leveldb.OpenFile(cfg.DataDir+"/blockchain/utxo", nil)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return &mgr
}

func (um *utxoMgr) manageUtxo() {
deadloop:
	for {
		select {
		case <-um.stop:
			break deadloop
		case tx := <-um.tx:
			log.Info("new tx to update uxto")
			var key [36]byte
			for i := uint64(0); i < tx.TxinCount.Value; i++ {
				//消耗UTXO,把leveldb中的UTXO删除
				if tx.CoinbaseTx() { //去掉coinbase的交易输入（无效交易输入）
					break
				}
				copy(key[:], tx.Txins[i].PreOut.Hash[:])
				var i2b4 [4]byte
				binary.LittleEndian.PutUint32(i2b4[:], tx.Txins[i].PreOut.Index)
				copy(key[32:], i2b4[:])
				err := um.dbUtxo.Delete(key[:], nil)
				if err != nil {
					panic(err)
				}
			}

			// 放到for循环外面优化处理速度
			txid := tx.Txid()
			copy(key[:], txid[:])
			for i := uint64(0); i < tx.TxoutCount.Value; i++ {
				//添加UTXO到leveldb中
				var i2b8 [8]byte
				binary.LittleEndian.PutUint64(i2b8[:], i)
				copy(key[32:], i2b8[:4]) //不会有那么大的索引
				err := um.dbUtxo.Put(key[:], tx.TxOuts[i].Serialize(), nil)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	um.dbUtxo.Close()
	close(um.stop)
	close(um.tx)
	log.Info("exit utxo manager...")
	um.done <- true
}

func startUtxoMgr(cfg *common.Config) {
	defaultUtxoMgr = newUtxoMgr(cfg)
	go defaultUtxoMgr.manageUtxo()
}

func stopUtxoMgr() {
	defaultUtxoMgr.stop <- true

	<-defaultUtxoMgr.done
	close(defaultUtxoMgr.done)
}

func utxo(key [36]byte) (*p2p.TxOutput, error) {
	var buf []byte
	var err error
	if buf, err = defaultUtxoMgr.dbUtxo.Get(key[:], nil); err != nil {
		return nil, err
	}
	output := p2p.TxOutput{}
	if err = output.Parse(buf); err != nil {
		return nil, err
	}

	return &output, nil
}

// 查询一个input的金额
func inputValue(preOut p2p.PreOutput) uint64 {
	// 从leveldb里面查
	var zeroHash [32]byte
	if reflect.DeepEqual(zeroHash[:], preOut.Hash[:]) {
		return uint64(0)
	}
	var key [36]byte
	copy(key[:], preOut.Hash[:])
	var i2b4 [4]byte
	binary.LittleEndian.PutUint32(i2b4[:], preOut.Index)
	copy(key[32:], i2b4[:])
	txout, err := utxo(key)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return txout.Value
}

// 计算n笔交易的手续费之和
func Fee(txs []p2p.TxPayload) uint64 {
	// 分别统计所有Vin的值和Vout的值，做二者的差值就是费用
	var inValueSum, outValueSum uint64
	for i := 0; i < len(txs); i++ {
		for j := uint64(0); j < txs[i].TxinCount.Value; j++ {
			inValueSum += inputValue(txs[i].Txins[j].PreOut) //
		}

		for j := uint64(0); j < txs[i].TxoutCount.Value; j++ {
			outValueSum += txs[i].TxOuts[j].Value
		}
	}

	return outValueSum - inValueSum
}
