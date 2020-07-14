package storage

import (
	"btcnetwork/common"
	"btcnetwork/p2p"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ErrBlockNotFound = errors.New("block not found")
	ErrTxNotFound    = errors.New("tx not found")
	ErrUtxoNotFound  = errors.New("UTXO not found")
)

var log *logrus.Logger

func Store(newBlock *p2p.BlockPayload) {
	defaultBlockMgr.newBlock <- *newBlock
}

func Start(cfg *common.Config) {
	startBlockMgr(cfg)
	startTxMgr(cfg)
	startUtxoMgr(cfg)
}

func Stop() {
	stopBlockMgr()
	stopTxMgr()
	stopUtxoMgr()
}

// todo:根据区块哈希找出区块数据
func BlockFromHash(hash [32]byte) (*p2p.BlockPayload, error) {
	return nil, ErrBlockNotFound
}

func HasBlockHash(hash [32]byte) bool {
	has, err := defaultBlockMgr.DBhash2block.Has(hash[:], nil)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return has
}

func LatestBlockHeight() uint32 {
	if defaultBlockMgr.IsEmpty() { //如果是空的就返回创世区块哈希
		return 0
	}
	var buf []byte
	var err error
	if buf, err = defaultBlockMgr.DBlatestblock.Get(LatestBlockKey, nil); err != nil {
		log.Error(err)
		panic(err)
	}
	return binary.LittleEndian.Uint32(buf)
}

func LatestBlockHash() [32]byte {
	if defaultBlockMgr.IsEmpty() { //如果是空的就返回创世区块哈希
		return defaultBlockMgr.genesisBlockHash()
	}
	var buf []byte
	var err error
	if buf, err = defaultBlockMgr.DBlatestblock.Get(LatestBlockKey, nil); err != nil {
		log.Error(err)
		panic(err)
	}
	var hash [32]byte
	if buf, err = defaultBlockMgr.DBheight2hash.Get(buf, nil); err != nil {
		log.Error(err)
		panic(err)
	}
	copy(hash[:], buf)
	return hash
}

// todo:根据区块高度找出区块数据
func BlockFromHeight(hash [32]byte) (*p2p.BlockPayload, error) {
	return nil, ErrBlockNotFound
}

// todo:根据交易交易id找出交易数据
func Tx(txid [32]byte) (*p2p.TxPayload, error) {
	return nil, ErrTxNotFound
}

// 根据PreOut组成的key找出交易输出数据
//func Utxo(key [36]byte) (*p2p.TxOutput, error) {
//	txout,err := utxo(key)
//	if err != nil {
//		log.Error(err)
//		return nil,ErrUtxoNotFound
//	}
//	return txout, nil
//}

func init() {
	log = logrus.New()
}
