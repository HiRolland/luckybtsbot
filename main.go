package main

import (
	"syscall"

	"github.com/vrecan/death"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/updater"
	"github.com/zhangpanyi/botcasino/config"
	"github.com/zhangpanyi/botcasino/logic"
	context "github.com/zhangpanyi/botcasino/logic/context"
	"github.com/zhangpanyi/botcasino/logic/notice"
	"github.com/zhangpanyi/botcasino/logic/syncfee"
	"github.com/zhangpanyi/botcasino/logic/timer"
	"github.com/zhangpanyi/botcasino/models"
	"github.com/zhangpanyi/botcasino/pusher"
	"github.com/zhangpanyi/botcasino/storage"
	"github.com/zhangpanyi/botcasino/webhook"
	"github.com/zhangpanyi/botcasino/webrpc"
	"github.com/zhangpanyi/botcasino/withdraw"
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

	// 创建更新器
	botUpdater, err := updater.NewUpdater(serveCfg.Port, serveCfg.Domain, serveCfg.APIWebsite)
	if err != nil {
		logger.Panic(err)
	}
	webrpc.InitRoute(botUpdater.GetRouter())
	webhook.InitRoute(botUpdater.GetRouter())

	// 同步转账手续费
	syncfee.CheckFeeStatusAsync()

	// 运行转账服务
	withdraw.RunWithdrawServiceForOnce(6)

	// 启动红包机器人
	context.CreateManagerForOnce(serveCfg.BucketNum)
	bot, err := botUpdater.AddHandler(serveCfg.Token, logic.NewUpdate)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infof("Lucky money bot id is: %d", bot.ID)
	pool := updater.NewPool(2048)
	timer.StartTimerForOnce(bot, pool)

	// 初始化推送配置
	notice.InitBotForOnce(bot)

	// 创建消息推送器
	pusher.CreatePusherForOnce(pool)

	// 启动更新服务器
	logger.Infof("Lucky money server started")
	go func() {
		err = botUpdater.ListenAndServe()
		if err != nil {
			logger.Panicf("Lucky money server failed to listen: %v", err)
		}
	}()

	// 捕捉退出信号
	d := death.NewDeath(syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL,
		syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM)
	d.WaitForDeathWithFunc(func() {
		storage.Close()
		logger.Info("Lucky money server stoped")
	})
}
