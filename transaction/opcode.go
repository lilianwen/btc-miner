package transaction

type OpCode byte

const (
	OP_DUP         OpCode = 0x76 // 118
	OP_HASH160     OpCode = 0xa9 // 169
	OP_EQUALVERIFY OpCode = 0x88 // 136
	OP_CHECKSIG    OpCode = 0xac // 172
)
