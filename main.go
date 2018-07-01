package main

import (
	"net/http"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/vrecan/death"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/updater"
	"github.com/zhangpanyi/botcasino/app/admin"
	"github.com/zhangpanyi/botcasino/app/config"
	"github.com/zhangpanyi/botcasino/app/httprpc"
	"github.com/zhangpanyi/botcasino/app/logic"
	context "github.com/zhangpanyi/botcasino/app/logic/context"
	"github.com/zhangpanyi/botcasino/app/logic/notice"
	"github.com/zhangpanyi/botcasino/app/logic/syncfee"
	"github.com/zhangpanyi/botcasino/app/logic/timer"
	"github.com/zhangpanyi/botcasino/app/models"
	"github.com/zhangpanyi/botcasino/app/poll"
	"github.com/zhangpanyi/botcasino/app/pusher"
	"github.com/zhangpanyi/botcasino/app/storage"
	"github.com/zhangpanyi/botcasino/app/webhook"
	"github.com/zhangpanyi/botcasino/app/withdraw"
	"upper.io/db.v3/sqlite"
)

func main() {
	// 加载配置文件
	config.LoadConfig("master.yml")

	// 初始化日志库
	serveCfg := config.GetServe()
	logger.CreateLoggerOnce(logger.DebugLevel, logger.InfoLevel)

	// 连接到数据库
	err := storage.Connect(serveCfg.BolTDBPath)
	if err != nil {
		logger.Panic(err)
	}
	dbcfg := serveCfg.SQlite
	settings := sqlite.ConnectionURL{
		Database: dbcfg.Database,
		Options:  dbcfg.Options,
	}
	err = models.Connect(settings)
	if err != nil {
		logger.Panic(err)
	}

	// 设置系统账户
	account, err := httprpc.GetAccount()
	if err != nil {
		logger.Panic(err)
	}
	config.SetAccount(account)

	// 同步转账手续费
	syncfee.CheckFeeStatusAsync()

	// 运行转账服务
	withdraw.RunWithdrawServiceForOnce(6)

	// 创建消息上下文管理
	context.CreateManagerForOnce(serveCfg.BucketNum)

	// 创建轮询器
	poller := poll.NewPoller(serveCfg.APIWebsite)
	bot, err := poller.StartPoll(serveCfg.Token, logic.NewUpdate)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infof("Lucky money bot id is: %d", bot.ID)

	// 初始化推送配置
	notice.InitBotForOnce(bot)

	// 启动红包定时器
	pool := updater.NewPool(2048)
	timer.StartTimerForOnce(bot, pool)

	// 创建消息推送器
	pusher.CreatePusherForOnce(pool)

	// 启动HTTP服务器
	router := mux.NewRouter()
	admin.InitRoute(router)
	webhook.InitRoute(router)
	addr := serveCfg.Host + ":" + strconv.Itoa(serveCfg.Port)
	go func() {
		s := &http.Server{
			Addr:    addr,
			Handler: router,
		}
		if err = s.ListenAndServe(); err != nil {
			logger.Panicf("Failed to listen and serve, %v, %v", addr, err)
		}
	}()
	logger.Infof("Lucky money server started")

	// 捕捉退出信号
	d := death.NewDeath(syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL,
		syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM)
	d.WaitForDeathWithFunc(func() {
		storage.Close()
		logger.Info("Lucky money server stoped")
	})
}
