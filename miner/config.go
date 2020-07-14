package miner

import "encoding/hex"

type Config struct {
	Version      uint32
	Target       [32]byte
	Bits         uint32
	CurrHeight   uint32
	PreBlockHash [32]byte
	//区块容量上限
	//区块奖励
	Reward          uint64
	MinerPubKeyHash [20]byte
}

var minerConfig = InitConfig()

//初始化配置信息要从区块0重放区块头进行计算
func InitConfig() *Config {
	cfg := Block0Config()
	//todo:还有一部分数据从外面输入，如矿工的地址
	// cfg.MinerPubKeyHash
	cfg = EvolveConfig(cfg)
	return cfg
}

// Evolve进化的意思
// 随着区块高度的增加，配置信息也可能会改变, 如难度值，区块奖励
func EvolveConfig(cfg *Config) *Config {
	//todo: 将来实现，暂时不考虑参数变化
	return cfg
}

//区块0的配置信息，写死在代码里
func Block0Config() *Config {
	cfg := Config{}
	cfg.Version = 0x01
	buf, _ := hex.DecodeString("00000000ffff0000000000000000000000000000000000000000000000000000")
	copy(cfg.Target[:], buf)
	buf, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	copy(cfg.PreBlockHash[:], buf)
	cfg.CurrHeight = 0
	cfg.Bits = 0x1d00ffff
	cfg.Reward = 50 * 100000000

	return &cfg
}
