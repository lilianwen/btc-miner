package miner

import (
	"btcnetwork/common"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Start(cfg *common.Config) {
	minerConfig = InitConfig(cfg)

}

func init() {
	log = logrus.New()
}
