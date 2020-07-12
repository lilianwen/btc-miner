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
