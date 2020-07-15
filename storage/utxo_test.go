package storage

import (
	"encoding/hex"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
)

func TestLoadAllUtxo(t *testing.T) {
	db, err := leveldb.OpenFile("F:/go/src/btcnetwork/data/blockchain/utxo", nil)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		fmt.Println(hex.EncodeToString(iter.Key()))
	}
	iter.Release()

	//err = db.Put([]byte("test"), []byte("yes"), nil)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//iter = db.NewIterator(nil, nil)
	//for iter.Next() {
	//	fmt.Println(hex.EncodeToString(iter.Key()))
	//}
	//iter.Release()
	//err = db.Delete([]byte("test"), nil)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

}
