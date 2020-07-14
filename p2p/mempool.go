package p2p

func NewMempoolMsg() (*Msg, error) {
	return NewMsg("mempool", nil)
}
