package groupchat

import (
	"fmt"
	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
)

// 获取群组ID
type GetChatIDHandler struct {
}

// 消息处理
func (handler *GetChatIDHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	fromID := update.Message.From.ID
	reply := tr(fromID, "lng_chat_get_chat_id")
	reply = fmt.Sprintf(reply, update.Message.Chat.ID, bot.UserName, bot.UserName)
	bot.ReplyMessageDisableWebPagePreview(update.Message, reply, true, nil)
}

// 消息路由
func (handler *GetChatIDHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}