package webhook

import (
	"sync"

	"github.com/gorilla/mux"
)

var once sync.Once

// 初始路由
func InitRoute(router *mux.Router) {
	once.Do(func() {
		router.HandleFunc("/bitshares/webhook", handleTransferOperation)
	})
}
