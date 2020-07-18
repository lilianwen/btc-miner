package node

import (
	"btcnetwork/common"
	"btcnetwork/storage"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

var (
//ErrorInnerErr = errors.Errorf("service inner error")
)

func (node *Node) apiMempool(w http.ResponseWriter, r *http.Request) {
	//读取当前mempool里的所有交易哈希值
	var txs []string
	var count = uint32(0)
	node.txPool.Range(func(key, value interface{}) bool {
		txs = append(txs, hex.EncodeToString(key.([]byte)))
		count++
		return true
	})

	buf, err := json.Marshal(txs)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err = w.Write(buf); err != nil {
		log.Error(err)
	}
}

func (node *Node) apiLatest(w http.ResponseWriter, r *http.Request) {
	latest := storage.LatestBlockHeight()
	buf, err := json.Marshal(latest)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err = w.Write(buf); err != nil {
		log.Error(err)
	}
}

// 挖一个区块
func (node *Node) apiMineOne(w http.ResponseWriter, r *http.Request) {
	common.MinerCmd <- common.MineOneTime
}

// 不停地自动挖矿
func (node *Node) apiAutoMine(w http.ResponseWriter, r *http.Request) {
	common.MinerCmd <- common.AutoMine
}

// 停止挖矿
func (node *Node) apiStopMine(w http.ResponseWriter, r *http.Request) {
	common.MinerCmd <- common.StopMine
}
