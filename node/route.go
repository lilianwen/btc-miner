package node

import (
	"net/http"
	"sync"
)

func (node *Node) StartApiService(wg *sync.WaitGroup) {
	defer wg.Done()
	var handlers = map[string]func(*Node, http.ResponseWriter, *http.Request){
		"/mempool":  (*Node).apiMempool,
		"/latest":   (*Node).apiLatest,
		"/mineone":  (*Node).apiMineOne,
		"/automine": (*Node).apiAutoMine,
		"/stopmine": (*Node).apiStopMine,
	}
	node.apiHandlers = handlers
	mux := http.NewServeMux()
	mux.Handle("/", *node)
	if err := http.ListenAndServe(node.Cfg.RpcListen, mux); err != nil {
		panic(err)
	}
}

func (node Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RequestURI)
	handler := node.apiHandlers[r.RequestURI]
	handler(&node, w, r)
}
