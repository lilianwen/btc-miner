package storage

import (
	"btcnetwork/common"
	"btcnetwork/p2p"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"reflect"
	"sync"
)

var (
	LatestBlockKey = []byte("latestblock")
	ErrOphanBlock  = errors.New("ophan block")
)

type blockMgr struct {
	newBlock       chan p2p.BlockPayload
	DBhash2block   *leveldb.DB
	DBhash2height  *leveldb.DB
	DBheight2hash  *leveldb.DB
	DBlatestblock  *leveldb.DB
}

var defaultBlockMgr *blockMgr

func startBlockMgr(cfg *common.Config, ctx context.Context, wg *sync.WaitGroup) {
	defaultBlockMgr = newBlockMgr(cfg)
	go defaultBlockMgr.manageBlockDB(ctx, wg)
}

func newBlockMgr(cfg *common.Config) *blockMgr {
	s := blockMgr{}
	s.newBlock = make(chan p2p.BlockPayload, 500) //todo:仔细考量一下这个数字该如何定

	var err error
	if s.DBhash2block, err = leveldb.OpenFile(cfg.DataDir+"/blockchain/block/hash2block", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	if s.DBhash2height, err = leveldb.OpenFile(cfg.DataDir+"/blockchain/block/hash2height", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	if s.DBheight2hash, err = leveldb.OpenFile(cfg.DataDir+"/blockchain/block/height2hash", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	if s.DBlatestblock, err = leveldb.OpenFile(cfg.DataDir+"/blockchain/block/latestblock", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	return &s
}

func (bm *blockMgr) manageBlockDB(ctx context.Context,wg *sync.WaitGroup) {
	nb := p2p.BlockPayload{}
deadloop:
	for {
		select {
		case <-ctx.Done():
			break deadloop

		case nb = <-bm.newBlock:
			log.Info("update a block...")
			err := bm.updateDBs(&nb)
			if err != nil {
				if err == ErrOphanBlock {
					log.Info(err)
					continue
				}
				log.Error(err)
				break deadloop
			}
			log.Info("update a block done.")
		}
	}
	_ = bm.DBhash2block.Close()
	_ = bm.DBheight2hash.Close()
	_ = bm.DBhash2height.Close()
	_ = bm.DBlatestblock.Close()
	close(bm.newBlock)
	log.Info("exit db mamager...")
	wg.Done()
}

//把新区块写入leveldb
func (bm *blockMgr) updateDBs(newBlock *p2p.BlockPayload) error {
	var preHeight uint32
	preHash := newBlock.PreHash
	genesisBlockHash := bm.genesisBlockHash()
	if reflect.DeepEqual(preHash[:], genesisBlockHash[:]) { //处理创世区块没有前区块高度的问题
		preHeight = 0
	} else {
		var err error
		preHeight, err = bm.hash2Height(preHash) //根据区块哈希找区块高度
		if err != nil {
			log.Error(err)
			//可能没找到这个哈希值，那当前这个区块可能就是一个孤块,处理孤块
			return ErrOphanBlock
		}
	}

	log.Debug("blockheader: >>", hex.EncodeToString(newBlock.Header.Serialize()))
	log.Debug("block: >>", hex.EncodeToString(newBlock.Serialize()))
	hash := common.Sha256AfterSha256(newBlock.Header.Serialize())
	log.Debug("storage----hash:", hex.EncodeToString(hash[:]))
	err := bm.DBhash2block.Put(hash[:], newBlock.Serialize(), nil)
	if err != nil {
		log.Error(err)
	}

	curHeight := preHeight + 1
	var i2b4 [4]byte
	binary.LittleEndian.PutUint32(i2b4[:], curHeight)
	if err = bm.DBheight2hash.Put(i2b4[:], hash[:], nil); err != nil {
		log.Error(err)
		return err
	}
	//
	if err = bm.DBhash2height.Put(hash[:], i2b4[:], nil); err != nil {
		log.Error(err)
		return err
	}
	if err = bm.DBlatestblock.Put(LatestBlockKey, i2b4[:], nil); err != nil {
		log.Error(err)
		return err
	}

	//UTXO分析
	for i := uint64(0); i < newBlock.TxnCount.Value; i++ {
		defaultUtxoMgr.tx <- newBlock.Txns[i]
	}

	log.Debug("update block height:", curHeight)
	return nil
}

// 从leveldb中查找区块哈希对应的区块高度
func (bm *blockMgr) hash2Height(hash [32]byte) (uint32, error) {
	buf, err := bm.DBhash2height.Get(hash[:], nil)
	if err != nil {
		return 0, err
	}
	height := binary.LittleEndian.Uint32(buf)
	return height, nil
}

func (bm *blockMgr) genesisBlockHash() [32]byte {
	var buf []byte
	buf, _ = hex.DecodeString(common.GenesisBlockHash)
	var hash [32]byte
	copy(hash[:], buf)
	return hash
}

func (bm *blockMgr) IsEmpty() bool {
	has, err := defaultBlockMgr.DBlatestblock.Has(LatestBlockKey, nil)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return !has
}
