package p2p

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"encoding/hex"
	"testing"
)

func TestBlockPayload_Parse(t *testing.T) {
	//DataInString := "03000030a2dad216e046040d0b4b03ed8d65667cef6642e58fdf471078dfc7dea0aab35160a8228f5bc802b23c8d2aac7b32026af0ba6f7d137468a6acdfff2e410c522b3fc80a5fffff7f20000000000201000000010000000000000000000000000000000000000000000000000000000000000000ffffffff17016c084d38ab7e8a6291a00b2f503253482f627463642fffffffff01e0f2052a010000001976a9147b9ab762d867fef558c26f8b14233e206386d56588ac0000000001000000010cb9551bfc3b772e0efa0817dc069df8694f756017373638a16e74854fc0be1b000000006b483045022100c5e444cfe7d13779a6a772ab9692aae87d8cf7a1e2fb80306e05a30ee4b16d1402207f56b99125ea7d8a50cbb6990f8e9b43088b4965d81af1cf6a9dbbef2e5e46960121033bee237c0e48aad2cc4411f7c51f424a94f063452ce32457d94be2d91e51f712ffffffff0200e1f505000000001976a91482d55da28ed20143e127b21c4aacc062424f46d388ac20101024010000001600147f3f76a54d863d17b3533deaf697fe34e277cf7b00000000"
	DataInString := "0300003016f47d75475a67888235ea805ae59251766b37b1226e0cac2bca95463307bb1a545ce22a4b21d3e1a2c7876deba2f378e8f3ed378c29df871ec5a7da78578ebd45360b5fffff7f20000000000201000000010000000000000000000000000000000000000000000000000000000000000000ffffffff17017d089efec3ecbe8362b20b2f503253482f627463642fffffffff01e0f2052a010000001976a9147b9ab762d867fef558c26f8b14233e206386d56588ac0000000001000000016bf39aca3c7340115380760034795779cfedd8d83565359bcc72bdb994d72911000000006a473044022072e80628c7be07560c0e82d27a41edb28d067b6c7efc2c568a7dc17bc84ef222022070e875918f12217d2299bb9b30913718c28db22cbeaf748b88615a2861260d6e0121033bee237c0e48aad2cc4411f7c51f424a94f063452ce32457d94be2d91e51f712ffffffff0200e1f505000000001976a91482d55da28ed20143e127b21c4aacc062424f46d388ac20101024010000001600149c61159dfa36778091e2ce991640d6eba27ec9ae00000000"

	var data []byte
	data, err := hex.DecodeString(DataInString)
	if err != nil {
		t.Error(err)
	}
	recvBlock := BlockPayload{}
	err = recvBlock.Parse(data)
	if err != nil {
		t.Error(err)
	}

	var hashes []string
	for i := range recvBlock.Txns {
		txHash := recvBlock.Txns[i].TxHash()
		log.Infof("tx[%d].txhash=%s, size=%d", i, hex.EncodeToString(txHash[:]), recvBlock.Txns[i].Len())
		txid := recvBlock.Txns[i].Txid()
		log.Infof("tx[%d].txid=%s", i, hex.EncodeToString(txid[:]))
		hashes = append(hashes, hex.EncodeToString(common.ReverseBytes(txid[:])))
	}
	//尝试构建默克尔树
	root, err := block.ConstructMerkleRoot(hashes)
	wantMerkleRootHash := hex.EncodeToString(common.ReverseBytes(recvBlock.MerkleRootHash[:]))
	if root.Value != wantMerkleRootHash {
		log.Error("calculate merkle root hash not equal to block header merkle root hash")
		log.Errorf("get:%s, want:%s", root.Value, wantMerkleRootHash)
	}
}
