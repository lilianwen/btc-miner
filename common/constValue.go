package common

const (
	GenesisBlockHash   = "f67ad7695d9b662a72ff3d8edbbb2de0bfa67b13974bb9910d116d5cbd863e68" //simBlock
	GenesisBlockReward = 50 * 100000000
	GenesisBlockBits   = uint32(0x207fffff)                                                 //主网参数uint32(0x1d00ffff)
	GenesisTarget      = "7fffff0000000000000000000000000000000000000000000000000000000000" //主网参数"00000000ffff0000000000000000000000000000000000000000000000000000"
	GenesisBlockHeight = 0
	MinerVersion       = uint32(0x01) //挖矿协议的版本
)
