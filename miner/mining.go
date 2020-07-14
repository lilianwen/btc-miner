package miner

import (
	"btcnetwork/block"
	"btcnetwork/common"
	"btcnetwork/p2p"
	"btcnetwork/transaction"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"time"
)

var (
	ErrNonceNotFound = errors.New("nonce not found")
)

func Mining() {
	//获取交易
	txString := "01000000012208792e058abb8e0ba508c9c04bf93b28a1b28271bf127ed98435602603d287000000006b483045022100bdb8185a0bb9d977f1d6e9e0c470562abbfdebdd01e842671ae22abf2107b92402203a65644f4f88651350e3a199c65e092f8a58922152aeca9b66d1632da2773dad0121033bee237c0e48aad2cc4411f7c51f424a94f063452ce32457d94be2d91e51f712ffffffff0200e1f505000000001976a91482d55da28ed20143e127b21c4aacc062424f46d388ac2010102401000000160014454dad4856e0e2bd32b21eba992011132e29f2b000000000"
	buf, _ := hex.DecodeString(txString)
	tx := p2p.TxPayload{}
	_ = tx.Parse(buf)

	var txs []p2p.TxPayload
	txs = append(txs, tx)

	//构建coinbase交易
	var txids []string
	coinbase := p2p.TxPayload{}
	coinbase.Version = minerConfig.Version
	coinbase.Marker = nil
	coinbase.Flag = nil
	coinbase.TxinCount = common.NewVarInt(1)
	input := p2p.TxInput{}
	input.PreOut = p2p.NewCoinPreOutput()
	//var sigScript = p2p.
	//input.SigScript = //transaction.NewP2PKHScipt(minerConfig.MinerPubKeyHash)
	input.ScriptLen = common.NewVarInt(uint64(len(input.SigScript)))
	input.Sequence = 0xffffffff
	coinbase.Txins = append(coinbase.Txins, input)
	coinbase.TxoutCount = common.NewVarInt(1)
	output := p2p.TxOutput{}
	output.Value = minerConfig.Reward + p2p.Fee(txs) //由2部分构成，一部分是系统奖励，另一部分是交易手续费
	output.PkScript = transaction.NewP2PKHScipt(minerConfig.MinerPubKeyHash[:])
	output.PkScriptLen = common.NewVarInt(uint64(len(output.PkScript)))
	coinbase.TxOuts = append(coinbase.TxOuts, output)
	coinbase.WitnessCount = nil
	coinbase.TxWitnesses = nil
	coinbase.Locktime = 0

	//构建区块头挖矿
	txid := coinbase.Txid()
	txids = append(txids, hex.EncodeToString(txid[:]))
	root, _ := block.ConstructMerkleRoot(txids)
	buf, err := hex.DecodeString(root.Value)
	if err != nil {
		log.Errorf("merkle root hash decode error:%s", root.Value)
	}

	header := block.Header{}
	header.BlockVersion = 0x01
	header.PreHash = minerConfig.PreBlockHash
	copy(header.MerkleRootHash[:], buf) //注意：这里的值要不要逆序？
	header.Timestamp = uint32(time.Now().Second())
	header.Bits = minerConfig.Bits
	nonce, err := mine(&header, 0)
	if err != nil {
		//前期挖矿失败，就不挖了，todo:后期考虑调整交易顺序，使用扩展nonce等策略
		log.Error("not found avaliable nonce")
	} else {
		header.Nonce = nonce
		//组建区块，广播给其他节点

		//刷新minerConfig,每2016个区块调整难度值，可能需要调整难度值
		log.Infof("generate block: nonce=%d", nonce)
	}
}

func mine(header *block.Header, startNonce uint32) (uint32, error) {
	for i := startNonce; true; i++ {
		header.Nonce = i
		buf := header.Serialize()
		blockHash := common.Sha256AfterSha256(buf)

		target := block.Bits2Target(header.Bits)
		common.ReverseBytes(blockHash[:]) //注意：这里一定要反转一下顺序,因为big.Int是大端存储
		gotHash := new(big.Int).SetBytes(blockHash[:])
		if target.Cmp(gotHash) >= 0 { //bingo 挖到区块了
			return i, nil
		}
		//fmt.Println(i)
		if i == math.MaxUint32 {
			break
		}
	}
	return 0, ErrNonceNotFound
}
