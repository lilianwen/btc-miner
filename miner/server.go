package miner

import (
	"btcnetwork/common"
	"sync"
)

var wg sync.WaitGroup
func Start(cfg *common.Config) {
	log.Info("start miner service")
	minerConfig = InitConfig(cfg)
	minerStop = make(chan bool, 1)
	wg.Add(1)
	go mineMonitor(&wg)
}

func Stop() {
	log.Info("stop miner service")
	common.MinerCmd <- common.StopMine //这里无法用ctx.Done()替换，因为涉及到挖矿的三个状态，而不仅仅是是否在挖矿两个状态

	wg.Wait()
}
