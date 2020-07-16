package common

import "github.com/sirupsen/logrus"

var LogLevel = logrus.DebugLevel

type MineCmd uint32

const (
	StopMine    = MineCmd(0)
	MineOneTime = MineCmd(1)
	AutoMine    = MineCmd(2)
)

var MinerCmd = make(chan MineCmd, 1)
