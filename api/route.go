package api

import (
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func StartApiService() {
	//mux := http.NewServeMux()
	//mux.Handle("/mempool", MempoolHandler)
	//if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
	//	panic(err)
	//}
}

func init() {
	log = logrus.New()
}
