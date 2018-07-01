package groupchat

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/zhangpanyi/basebot/logger"
	tg "github.com/zhangpanyi/basebot/telegram"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/botcasino/app/config"
	"github.com/zhangpanyi/botcasino/app/storage"
)

// 随机器
var randx = rand.New(rand.NewSource(time.Now().UnixNano()))

// 获取广告
func getAd(botID int64) string {
	handler := storage.AdStorage{}
	ads, err := handler.GetAds(botID)
	if err != nil || len(ads) == 0 {
		return ""
	}
	ad := ads[randx.Intn(len(ads))]
	return fmt.Sprintf("\n\n*[* %s *]*", tg.Pre(ad.Text))
}

// 语言翻译
func tr(userID int64, key string) string {
	return config.GetLanguge().Value("zh_CN", key)
}

// 发送红包到群组
func SendLuckyMoneyToGroup(bot *methods.BotExt, userID, chatID int64, id uint64) error {
	// 获取红包信息
	newHandler := storage.LuckyMoneyStorage{}
	luckyMoney, received, err := newHandler.GetLuckyMoney(id)
	if err != nil {
		logger.Errorf("Failed to get lucky money, %v", err)
		return err
	}

	// 红包身份验证
	if userID != luckyMoney.SenderID {
		return errors.New("auth failed")
	}

	// 检查重复激活
	if luckyMoney.Active {
		return errors.New("auth failed")
	}

	// 检查红包过期
	now := time.Now().UTC().Unix()
	dynamicCfg := config.GetDynamic()
	if now-luckyMoney.Timestamp >= dynamicCfg.LuckyMoneyExpire {
		return errors.New("already activated")
	}

	// 生成菜单列表
	data := strconv.FormatUint(luckyMoney.ID, 10)
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{Text: tr(0, "lng_chat_receive"), CallbackData: data},
	}

	// 回复红包信息
	reply := tr(0, "lng_chat_welcome")
	typ := luckyMoneysTypeToString(luckyMoney.Lucky)
	amount := fmt.Sprintf("%.2f", float64(luckyMoney.Amount)/100.0)
	if !luckyMoney.Lucky {
		amount = fmt.Sprintf("%.2f", float64(luckyMoney.Amount*luckyMoney.Number)/100.0)
	}
	reply = fmt.Sprintf(reply, typ, received, luckyMoney.Number, amount,
		storage.GetAsset(luckyMoney.Asset), luckyMoney.SenderName,
		luckyMoney.Memo, getAd(bot.ID), bot.UserName, luckyMoney.ID, bot.UserName, luckyMoney.ID)
	markup := methods.MakeInlineKeyboardMarkup(menus[:], 1)
	message, err := bot.SendMessageDisableWebPagePreview(chatID, reply, true, markup)
	if err != nil {
		logger.Errorf("Failed to send lucky money info, %v", err)
		return err
	}

	// 激活红包
	err = newHandler.ActiveLuckyMoney(id, userID, message.Chat.ID, message.MessageID)
	if err != nil {
		logger.Errorf("Failed to active lucky money, %v", err)
		return err
	}
	return nil
}
