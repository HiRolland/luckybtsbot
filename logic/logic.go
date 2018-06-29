package logic

import (
	"github.com/zhangpanyi/botcasino/config"
	"github.com/zhangpanyi/botcasino/logic/context"
	"github.com/zhangpanyi/botcasino/logic/groupchat"
	"github.com/zhangpanyi/botcasino/logic/privatechat"
	"github.com/zhangpanyi/botcasino/storage"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
)

// 机器人更新
func NewUpdate(bot *methods.BotExt, update *types.Update) {
	// 是否暂停服务
	dynamicCfg := config.GetDynamic()
	if dynamicCfg.Suspended {
		return
	}

	// 获取用户ID
	var fromID int64
	var chatType string
	if update.Message != nil {
		fromID = update.Message.From.ID
		chatType = update.Message.Chat.Type
	} else if update.CallbackQuery != nil {
		fromID = update.CallbackQuery.From.ID
		chatType = update.CallbackQuery.Message.Chat.Type
	} else {
		logger.Debugf("Lucky money bot update, update_id: %v", update.UpdateID)
		return
	}
	logger.Debugf("Lucky money bot update, update_id: %v, user_id: %v", update.UpdateID, fromID)

	// 获取操作记录
	r, err := context.GetRecord(uint32(fromID))
	if err != nil {
		logger.Warnf("Failed to get bot record, bot_id: %v, %v, %v", bot.ID, fromID, err)
		return
	}

	// 添加机器人订户
	handler := storage.SubscriberStorage{}
	handler.AddSubscriber(bot.ID, fromID)

	// 处理机器人请求
	if chatType == types.ChatPrivate {
		new(privatechat.MainMenuHandler).Handle(bot, r, update)
	} else {
		new(groupchat.MainMenuHandler).Handle(bot, r, update)
	}

	// 删除空操作记录
	if r.Empty() {
		context.DelRecord(uint32(fromID))
	}
}
