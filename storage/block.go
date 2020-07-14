package storage

import (
	"btcnetwork/common"
	"btcnetwork/p2p"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
)

type blockMgr struct {
	stop          chan bool
	done          chan bool
	newBlock      chan p2p.BlockPayload
	DBhash2block  *leveldb.DB
	DBhash2height *leveldb.DB
	DBheight2hash *leveldb.DB
	DBlatestblock *leveldb.DB
}

var defaultBlockMgr *blockMgr

func startBlockMgr(cfg *common.Config) {
	defaultBlockMgr = newBlockMgr(cfg)
	go defaultBlockMgr.manageBlockDB()
}

func stopBlockMgr() {
	close(defaultBlockMgr.newBlock)
	defaultBlockMgr.stop <- true

	<-defaultBlockMgr.done
	close(defaultBlockMgr.done)
}

func newBlockMgr(cfg *common.Config) *blockMgr {
	s := blockMgr{}
	s.stop = make(chan bool, 1)
	s.done = make(chan bool, 1)
	s.newBlock = make(chan p2p.BlockPayload)

	//todo:很奇怪啊，这里必须先声明，不能用:=,以后在看看这是问么原因
	var err error
	s.DBhash2block, err = leveldb.OpenFile(cfg.DataDir+"/blockchain/block/hash2block", nil)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	if s.DBhash2height, err = leveldb.OpenFile(cfg.DataDir+"blockchain/block/hash2height", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	if s.DBheight2hash, err = leveldb.OpenFile(cfg.DataDir+"blockchain/block/height2hash", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	if s.DBlatestblock, err = leveldb.OpenFile(cfg.DataDir+"blockchain/block/latestblock", nil); err != nil {
		log.Error(err)
		panic(err)
	}

	return &s
}

func (bm *blockMgr) manageBlockDB() {
	nb := p2p.BlockPayload{}
deadloop:
	for {
		select {
		case <-bm.stop:
			break deadloop

		case nb = <-bm.newBlock:
			err := bm.updateDBs(&nb)
			if err != nil {
				break deadloop
			}
		}
	}
	_ = bm.DBhash2block.Close()
	_ = bm.DBheight2hash.Close()
	_ = bm.DBhash2height.Close()
	close(bm.newBlock)
	close(bm.stop)
	log.Info("exit db mamager")
	close(bm.done)
}

//把新区块写入leveldb
func (bm *blockMgr) updateDBs(newBlock *p2p.BlockPayload) error {
	hash := common.Sha256AfterSha256(newBlock.Header.Serialize())
	err := bm.DBhash2block.Put(hash[:], newBlock.Serialize(), nil)
	if err != nil {
		log.Error(err)
	}
	preHash := newBlock.PreHash
	height, err := bm.hash2Height(preHash) //根据区块哈希找区块高度
	if err != nil {
		//可能没找到这个哈希值，那当前这个区块可能就是一个孤块,处理孤块
	}
	curHeight := height + 1
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
	if err = bm.DBlatestblock.Put([]byte("latestblock"), i2b4[:], nil); err != nil {
		log.Error(err)
		return err
	}
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
