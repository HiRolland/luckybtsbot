package admin

import (
	"sync"

	"github.com/gorilla/mux"
	"github.com/zhangpanyi/botcasino/app/admin/handlers"
)

var once sync.Once

// 初始路由
func InitRoute(router *mux.Router) {
	once.Do(func() {
		router.HandleFunc("/admin/addad", handlers.AddAd)
		router.HandleFunc("/admin/delad", handlers.DelAd)
		router.HandleFunc("/admin/getads", handlers.GetAds)
		router.HandleFunc("/admin/addasset", handlers.AddAsset)
		router.HandleFunc("/admin/backup", handlers.Backup)
		router.HandleFunc("/admin/broadcast", handlers.Broadcast)
		router.HandleFunc("/admin/deductasset", handlers.DeductAsset)
		router.HandleFunc("/admin/frozen", handlers.Frozen)
		router.HandleFunc("/admin/unfrozen", handlers.Unfrozen)
		router.HandleFunc("/admin/get_assets", handlers.GetAssets)
		router.HandleFunc("/admin/restore", handlers.RestoreOrder)
		router.HandleFunc("/admin/subscribers", handlers.GetSubscribers)
	})
}
