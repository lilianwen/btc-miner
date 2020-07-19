package node

import (
	"context"
	"net/http"
	"sync"
)

func (node *Node) StartApiService(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var handlers = map[string]func(*Node, http.ResponseWriter, *http.Request){
		"/mempool":  (*Node).apiMempool,
		"/latest":   (*Node).apiLatest,
		"/mineone":  (*Node).apiMineOne,
		"/automine": (*Node).apiAutoMine,
		"/stopmine": (*Node).apiStopMine,
	}
	node.apiHandlers = handlers

	srv := &http.Server{
		Addr: node.Cfg.RpcListen,
		Handler: node,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Info("Httpserver: ListenAndServe() returns: ", err)
		}
	}()

	<- ctx.Done()
	if err := srv.Shutdown(nil); err != nil {
		log.Error(err)
	}
}

func (node Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RequestURI)
	handler := node.apiHandlers[r.RequestURI]
	handler(&node, w, r)
}
