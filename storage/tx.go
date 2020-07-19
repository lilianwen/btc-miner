package storage

import (
	"btcnetwork/common"
	"context"
	"sync"
)

//暂时用不上就不实现了
func startTxMgr(cfg *common.Config, ctx context.Context, wg *sync.WaitGroup) {
	_ = cfg
	_ = ctx
	wg.Done()
}
