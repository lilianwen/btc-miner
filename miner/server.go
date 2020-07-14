package miner

import "github.com/sirupsen/logrus"

var log *logrus.Logger

func Start() {
	InitConfig()

}

func init() {
	log = logrus.New()
}
