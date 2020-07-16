package miner

import (
	"btcnetwork/common"
	"encoding/hex"
)

type MiningState uint32

const (
	StateStop     = MiningState(0)
	StateOneBlock = MiningState(1)
	StateAuto     = MiningState(2)
)

type Config struct {
	Version    uint32
	Target     [32]byte
	Bits       uint32
	CurrHeight uint32
	//区块容量上限
	//区块奖励
	Reward          uint64
	MinerPubKeyHash [20]byte
	state           MiningState
}

var (
	minerConfig *Config
	//MineOneBlock chan bool
	//MineAuto chan bool

	minerStop chan bool
	Banner    = "lilianwen Mined"
)

//初始化配置信息要从区块0重放区块头进行计算
func InitConfig(cfg *common.Config) *Config {
	//把地址对应的公钥哈希计算出来
	addr, err := common.Base58Decode(cfg.MinerAddr)
	if err != nil {
		panic(err)
	}
	minerCfg := Block1Config()
	copy(minerCfg.MinerPubKeyHash[:], addr[1:21])
	minerCfg = EvolveConfig(minerCfg)
	minerCfg.state = StateStop
	return minerCfg
}

// Evolve进化的意思
// 随着区块高度的增加，配置信息也可能会改变, 如难度值，区块奖励
func EvolveConfig(cfg *Config) *Config {
	//todo: 将来实现，暂时不考虑参数变化
	// 需要根据当前区块高度调整bits值等信息
	return cfg
}

// 区块1的配置信息，根据创世区块相关信息填写
// 挖矿是从区块1开始的，区块0不用挖矿，区块0是写死在代码里的
func Block1Config() *Config {
	cfg := Config{}
	cfg.Version = common.MinerVersion
	buf, _ := hex.DecodeString(common.GenesisTarget)
	copy(cfg.Target[:], buf)
	cfg.CurrHeight = common.GenesisBlockHeight + 1
	cfg.Bits = common.GenesisBlockBits
	cfg.Reward = common.GenesisBlockReward

	return &cfg
}
