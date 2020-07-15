package storage

import (
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
)

func TestShowBlock(t *testing.T) {
	var (
		height2hashDB *leveldb.DB
		hash2blockDB  *leveldb.DB
		err           error
		hash          []byte
		blockRaw      []byte
	)
	if height2hashDB, err = leveldb.OpenFile("F:/go/src/btcnetwork/data/blockchain/block/height2hash", nil); err != nil {
		t.Error(err)
		return
	}
	if hash2blockDB, err = leveldb.OpenFile("F:/go/src/btcnetwork/data/blockchain/block/hash2block", nil); err != nil {
		t.Error(err)
		return
	}
	if hash, err = height2hashDB.Get([]byte{0x8a, 0x00, 0x00, 0x00}, nil); err != nil {
		t.Error(err)
		return
	}
	if blockRaw, err = hash2blockDB.Get(hash, nil); err != nil {
		t.Error(err)
		return
	}

	t.Log(hex.EncodeToString(blockRaw))
}
