package p2p

import (
	"net/http"
	"sync"
)

func (node *Node) StartApiService(wg *sync.WaitGroup) {
	defer wg.Done()
	mux := http.NewServeMux()
	mux.Handle("/mempool", *node)
	if err := http.ListenAndServe(node.Cfg.RpcListen, mux); err != nil {
		panic(err)
	}
}
