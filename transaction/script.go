package transaction

type ScriptBuilder struct {
	script []byte
	err    error
}

func NewScriptBuilder() *ScriptBuilder {
	return &ScriptBuilder{}
}

func (sb *ScriptBuilder) AddCode(code OpCode) *ScriptBuilder {
	sb.script = append(sb.script, byte(code))
	return sb
}

func (sb *ScriptBuilder) AddData(data []byte) *ScriptBuilder {
	sb.script = append(sb.script, data...)
	return sb
}

func (sb *ScriptBuilder) Script() []byte {
	return sb.script
}

func NewP2PKHScipt(pubKeyHash []byte) []byte {
	builder := NewScriptBuilder()
	builder.AddCode(OP_DUP).AddCode(OP_HASH160)
	builder.AddData(pubKeyHash)
	builder.AddCode(OP_EQUALVERIFY).AddCode(OP_CHECKSIG)
	return builder.Script()
}
