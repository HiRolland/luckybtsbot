package timer

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/zhangpanyi/botcasino/app/config"
	"github.com/zhangpanyi/botcasino/app/logic/groupchat"
	"github.com/zhangpanyi/botcasino/app/models"
	"github.com/zhangpanyi/botcasino/app/storage"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/updater"
)

var once sync.Once
var globalExpireTimer *expireTimer

// 启动定时器
func StartTimerForOnce(bot *methods.BotExt, pool *updater.Pool) {
	once.Do(func() {
		// 获取最后过期红包
		handler := storage.LuckyMoneyStorage{}
		id, err := handler.GetLastExpired()
		if err != nil && err != storage.ErrNoBucket {
			logger.Panic(err)
		}

		// 遍历未过期列表
		h := make(expireHeap, 0)
		err = handler.ForeachLuckyMoney(id+1, func(data *storage.LuckyMoney) {
			heap.Push(&h, expire{ID: data.ID, Timestamp: data.Timestamp})
		})
		if err != nil && err != storage.ErrNoBucket {
			logger.Panic(err)
		}

		// 初始化过期定时器
		globalExpireTimer = &expireTimer{
			h:    h,
			bot:  bot,
			pool: pool,
		}
		go globalExpireTimer.loop()
	})
}

// 获取机器人
func GetBot() *methods.BotExt {
	return globalExpireTimer.bot
}

// 添加红包
func AddLuckyMoney(id uint64, timestamp int64) {
	globalExpireTimer.lock.Lock()
	defer globalExpireTimer.lock.Unlock()
	heap.Push(&globalExpireTimer.h, expire{ID: id, Timestamp: timestamp})
}

// 过期定时器
type expireTimer struct {
	h    expireHeap
	bot  *methods.BotExt
	pool *updater.Pool
	lock sync.RWMutex
}

// 处理过期红包
func (t *expireTimer) handleLuckyMoneyExpire() {
	now := time.Now().UTC().Unix()
	dynamicCfg := config.GetDynamic()

	var id uint64
	t.lock.RLock()
	for t.h.Len() > 0 {
		data := t.h.Front()
		t.lock.RUnlock()

		// 判断是否过期
		if now-data.Timestamp < dynamicCfg.LuckyMoneyExpire {
			return
		}

		// 获取过期信息
		t.lock.Lock()
		e := heap.Pop(&t.h).(expire)
		t.lock.Unlock()

		id = e.ID
		logger.Infof("Lucky money expired, %v", e.Timestamp)
		t.pool.Async(func() {
			t.handleLuckyMoneyExpireAsync(e.ID)
		})
		t.lock.RLock()
	}
	t.lock.RUnlock()

	// 更新过期红包
	if id != 0 {
		handler := storage.LuckyMoneyStorage{}
		if err := handler.SetLastExpired(id); err != nil {
			logger.Warnf("Failed to set last expired  of lucky money, %v", err)
		}
	}
}

// 异步处理过期红包
func (t *expireTimer) handleLuckyMoneyExpireAsync(id uint64) {
	// 设置红包过期
	handler := storage.LuckyMoneyStorage{}
	if handler.IsExpired(id) {
		return
	}
	err := handler.SetExpired(id)
	if err != nil {
		logger.Infof("Failed to set expired of lucky money, %v", err)
		return
	}

	// 获取红包信息
	luckyMoney, received, err := handler.GetLuckyMoney(id)
	if err != nil {
		logger.Warnf("Failed to set expired of lucky money, not found lucky money, %d, %v", id, err)
		return
	}
	if received == luckyMoney.Number {
		return
	}

	// 计算红包余额
	balance := luckyMoney.Amount - luckyMoney.Received
	if !luckyMoney.Lucky {
		balance = luckyMoney.Amount*luckyMoney.Number - luckyMoney.Received
	}

	// 返还红包余额
	assetHandler := storage.AssetStorage{}
	err = assetHandler.UnfreezeAsset(luckyMoney.SenderID, luckyMoney.Asset, balance)
	if err != nil {
		logger.Errorf("Failed to return lucky money asset of expired, %v", err)
	} else {
		logger.Errorf("Return lucky money asset of expired, UserID=%d, Asset=%s, Amount=%d",
			luckyMoney.SenderID, luckyMoney.Asset, balance)
		desc := fmt.Sprintf("您发放的红包(id: *%d*)过期无人领取, 退还余额*%.2f* *%s*", luckyMoney.ID,
			float64(balance)/100.0, luckyMoney.Asset)
		models.InsertHistory(luckyMoney.SenderID, desc)
	}

	// 更新聊天信息
	groupchat.UpdateLuckyMoney(t.bot, luckyMoney, received)
}

// 事件循环
func (t *expireTimer) loop() {
	tickTimer := time.NewTimer(time.Second)
	for {
		select {
		case <-tickTimer.C:
			t.handleLuckyMoneyExpire()
			tickTimer.Reset(time.Second)
		}
	}
}
