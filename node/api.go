package node

import (
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
	node.mu.RLock()
	for txHash := range node.txPool {
		txs = append(txs, hex.EncodeToString(txHash[:]))
		count++
	}
	node.mu.RUnlock()

	buf, err := json.Marshal(txs)
	if err != nil {
		log.Errorln(err)
		return
	}
	if _, err = w.Write(buf); err != nil {
		log.Errorln(err)
	}
}

func (node *Node) apiLatest(w http.ResponseWriter, r *http.Request) {
	latest := storage.LatestBlockHeight()
	buf, err := json.Marshal(latest)
	if err != nil {
		log.Errorln(err)
		return
	}
	if _, err = w.Write(buf); err != nil {
		log.Errorln(err)
	}
}
