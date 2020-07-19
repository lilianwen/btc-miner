package storage

import (
	"btcnetwork/common"
	"btcnetwork/p2p"
	"context"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
	"sync"
)

var (
	ErrBlockNotFound          = errors.New("block not found")
	ErrBlockUnserializeFailed = errors.New("block unserialize failed")
	//ErrTxNotFound             = errors.New("tx not found")
	//ErrUtxoNotFound           = errors.New("UTXO not found")
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
)

func Store(newBlock *p2p.BlockPayload) {
	if !isStop() { //修复在newBlock通道关闭的情况下继续往通道里写数据的bug
		defaultBlockMgr.newBlock <- *newBlock
	}
}

func StoreSync(newBlock *p2p.BlockPayload) error {
	return defaultBlockMgr.updateDBs(newBlock)
}

func Start(cfg *common.Config) {
	ctx, cancel = context.WithCancel(context.Background())

	wg.Add(1)
	startBlockMgr(cfg, ctx, &wg)

	wg.Add(1)
	startTxMgr(cfg, ctx, &wg)

	wg.Add(1)
	startUtxoMgr(cfg, ctx, &wg)
}

func Stop() {
	cancel()
	wg.Wait()
}

func BlockFromHash(hash [32]byte) (*p2p.BlockPayload, error) {
	log.Debug(hex.EncodeToString(hash[:]))
	buf, err := defaultBlockMgr.DBhash2block.Get(hash[:], nil)
	if err != nil {
		log.Error(err)
		return nil, ErrBlockNotFound
	}
	blk := p2p.BlockPayload{}
	if err = blk.Parse(buf); err != nil {
		log.Error(err)
		return nil, ErrBlockUnserializeFailed
	}
	return &blk, nil
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

func isStop() bool {
	select {
	case <- ctx.Done():
		return true
	default:
		return false
	}
}

