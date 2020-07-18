package storage

import "github.com/btcsuite/btclog"

var log btclog.Logger

func init() {
	log = btclog.Disabled
}

func UseLogger(logger btclog.Logger) {
	log = logger
}
