package privatechat

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/botcasino/app/logic/groupchat"
)

// 匹配红包ID
var reMathLuckyMoneyID *regexp.Regexp

func init() {
	var err error
	reMathLuckyMoneyID, err = regexp.Compile("^/chatid/(\\d+)/")
	if err != nil {
		panic(err)
	}
}

// 选择群组
type SelectGroupHandler struct {
}

// 消息处理
func (*SelectGroupHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 提示输入
	back, err := r.Back()
	fromID := update.CallbackQuery.From.ID
	if err != nil || back.Message == nil {
		reply := tr(fromID, "lng_chat_enter_chat_id_say")
		bot.AnswerCallbackQuery(update.CallbackQuery, "", false, "", 0)
		bot.SendMessage(fromID, reply, true, nil)
		r.Clear().Push(update)
		return
	}

	// 获取红包ID
	data := update.CallbackQuery.Data
	result := reMathLuckyMoneyID.FindStringSubmatch(data)
	if len(result) != 2 {
		return
	}
	id, err := strconv.ParseUint(result[1], 10, 64)
	if err != nil {
		r.Clear()
		reply := tr(fromID, "lng_chat_enter_chat_id_failed")
		bot.SendMessage(fromID, reply, true, nil)
		return
	}

	// 获取群组ID
	chatID, err := strconv.ParseInt(back.Message.Text, 10, 64)
	if err != nil {
		reply := tr(fromID, "lng_chat_enter_chat_id_failed")
		bot.SendMessage(fromID, reply, true, nil)
		return
	}

	// 发送红包到群组
	err = groupchat.SendLuckyMoneyToGroup(bot, fromID, chatID, id)
	if err != nil {
		reply := tr(fromID, "lng_chat_enter_chat_id_send_failed")
		bot.SendMessage(fromID, reply, true, nil)
		return
	}

	// 回复处理结果
	r.Clear()
	reply := tr(fromID, "lng_chat_enter_chat_id_send_success")
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_menu"),
			CallbackData: "/main/",
		},
	}
	markup := methods.MakeInlineKeyboardMarkup(menus[:], 1)
	bot.SendMessage(fromID, fmt.Sprintf(reply, chatID), false, markup)

	// 更新消息内容
	menus = [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_chat_already_sent_out"),
			CallbackData: "/main/",
		},
	}
	markup = methods.MakeInlineKeyboardMarkup(menus[:], 1)
	reply = tr(fromID, "lng_priv_give_created_enter_chat_id")
	bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, true, markup)
}

// 消息路由
func (*SelectGroupHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}
