package api

import (
	"github.com/pkg/errors"
	"net/http"
)

var (
	ErrorInnerErr = errors.Errorf("service inner error")
)

type MempoolHandler struct {
}

func (mh *MempoolHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("this is mempool handler"))
	if err != nil {
		//写错误码到日志
		log.Errorln(err)
		//w.Write()
	}
}
