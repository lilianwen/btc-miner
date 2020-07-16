package miner

import (
	"btcnetwork/common"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Start(cfg *common.Config) {
	log.Info("start miner service")
	minerConfig = InitConfig(cfg)
	minerStop = make(chan bool, 1)
	go mineMonitor()
}

func Stop() {
	log.Info("stop miner service")
	common.MinerCmd <- common.StopMine

	<-minerStop
}

func init() {
	log = logrus.New()
	log.SetLevel(common.LogLevel)
}
