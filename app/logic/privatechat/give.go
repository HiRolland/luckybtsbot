package privatechat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zhangpanyi/botcasino/app/config"
	"github.com/zhangpanyi/botcasino/app/logic/core"
	"github.com/zhangpanyi/botcasino/app/logic/timer"
	"github.com/zhangpanyi/botcasino/app/models"
	"github.com/zhangpanyi/botcasino/app/storage"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
)

// 匹配资产
var reMathGiveAsset *regexp.Regexp

// 匹配类型
var reMathGiveType *regexp.Regexp

// 匹配金额
var reMathGiveAmount *regexp.Regexp

// 匹配数量
var reMathGiveNumber *regexp.Regexp

func init() {
	var err error
	reMathGiveAsset, err = regexp.Compile("^/give/(rand|equal)/$")
	if err != nil {
		panic(err)
	}

	reMathGiveType, err = regexp.Compile("^/give/(rand|equal)/(\\w+)/$")
	if err != nil {
		panic(err)
	}

	reMathGiveAmount, err = regexp.Compile("^/give/(rand|equal)/(\\w+)/([0-9]+\\.?[0-9]*)/$")
	if err != nil {
		panic(err)
	}

	reMathGiveNumber, err = regexp.Compile("^/give/(rand|equal)/(\\w+)/([0-9]+\\.?[0-9]*)/(\\d+)/$")
	if err != nil {
		panic(err)
	}
}

var (
	// 随机红包
	randLuckyMoney = "rand"
	// 普通红包
	equalLuckyMoney = "equal"
)

// 红包类型转字符串
func luckyMoneysTypeToString(fromID int64, typ string) string {
	if typ == randLuckyMoney {
		return tr(fromID, "lng_priv_give_rand")
	}
	return tr(fromID, "lng_priv_give_equal")
}

// 发放红包
type GiveHandler struct {
}

// 红包信息
type luckyMoneys struct {
	typ    string // 红包类型
	asset  string // 资产类型
	amount uint32 // 红包金额
	number uint32 // 红包个数
	memo   string // 红包备注
}

// 消息处理
func (handler *GiveHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 处理选择资产
	data := update.CallbackQuery.Data
	if data == "/give/" {
		r.Clear()
		handler.handleChooseType(bot, update.CallbackQuery)
		return
	}

	// 处理选择类型
	info := luckyMoneys{}
	result := reMathGiveAsset.FindStringSubmatch(data)
	if len(result) == 2 {
		r.Clear()
		info.typ = result[1]
		handler.handleChooseAsset(bot, &info, update.CallbackQuery)
		return
	}

	// 处理红包金额
	result = reMathGiveType.FindStringSubmatch(data)
	if len(result) == 3 {
		info.typ = result[1]
		info.asset = result[2]
		handler.handleLuckyMoneyAmount(bot, r, &info, update)
		return
	}

	// 处理红包个数
	result = reMathGiveAmount.FindStringSubmatch(data)
	if len(result) == 4 {
		info.typ = result[1]
		info.asset = result[2]
		amount, _ := strconv.ParseFloat(result[3], 10)
		info.amount = uint32(amount * 100)
		handler.handleLuckyMoneyNumber(bot, r, &info, update, true)
		return
	}

	// 处理红包留言
	result = reMathGiveNumber.FindStringSubmatch(data)
	if len(result) == 5 {
		info.typ = result[1]
		info.asset = result[2]
		amount, _ := strconv.ParseFloat(result[3], 10)
		info.amount = uint32(amount * 100)
		number, _ := strconv.Atoi(result[4])
		info.number = uint32(number)
		handler.handleLuckyMoneyMemo(bot, r, &info, update)
		return
	}

	// 路由到其它处理模块
	newHandler := handler.route(bot, update.CallbackQuery)
	if newHandler == nil {
		return
	}
	newHandler.Handle(bot, r, update)
}

// 消息路由
func (handler *GiveHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}

// 返回上级
func backSuperior(data string) string {
	s := strings.Split(data, "/")
	if len(s) <= 2 {
		return "/main/"
	}
	return strings.Join(s[:len(s)-2], "/") + "/"
}

// 生成基本菜单
func makeGiveBaseMenus(fromID int64, data string) *methods.InlineKeyboardMarkup {
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_priv_give_cancel"),
			CallbackData: "/main/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(data),
		},
	}
	return methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
}

// 处理选择类型
func (handler *GiveHandler) handleChooseType(bot *methods.BotExt, query *types.CallbackQuery) {

	// 生成菜单列表
	data := query.Data
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_priv_give_rand"),
			CallbackData: data + randLuckyMoney + "/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_priv_give_equal"),
			CallbackData: data + equalLuckyMoney + "/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: "/main/",
		},
	}

	// 回复请求结果
	bot.AnswerCallbackQuery(query, "", false, "", 0)
	reply := tr(fromID, "lng_priv_give_choose_type")
	markup := methods.MakeInlineKeyboardMarkup(menus[:], 2, 1)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
}

// 处理选择资产
func (handler *GiveHandler) handleChooseAsset(bot *methods.BotExt, info *luckyMoneys, query *types.CallbackQuery) {
	// 生成菜单列表
	data := query.Data
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{Text: "💴 bitCNY", CallbackData: data + storage.BitCNY + "/"},
		methods.InlineKeyboardButton{Text: "💵 bitUSD", CallbackData: data + storage.BitUSD + "/"},
		methods.InlineKeyboardButton{Text: tr(fromID, "lng_priv_give_cancel"), CallbackData: "/main/"},
		methods.InlineKeyboardButton{Text: tr(fromID, "lng_back_superior"), CallbackData: backSuperior(data)},
	}

	// 获取资产信息
	bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
	bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)

	// 回复请求结果
	bot.AnswerCallbackQuery(query, "", false, "", 0)
	markup := methods.MakeInlineKeyboardMarkup(menus[:], 2, 1, 1)
	reply := fmt.Sprintf(tr(fromID, "lng_priv_give_choose_asset"), bitCNY, bitUSD,
		luckyMoneysTypeToString(fromID, info.typ))
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
}

// 处理输入红包金额
func (handler *GiveHandler) handleEnterLuckyMoneyAmount(bot *methods.BotExt, r *history.History,
	info *luckyMoneys, update *types.Update, enterAmount string) {

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID

	// 处理错误
	data := query.Data
	handlerError := func(reply string) {
		r.Pop()
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := makeGiveBaseMenus(fromID, query.Data)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 检查输入金额
	amount, err := strconv.ParseFloat(enterAmount, 10)
	if err != nil || amount < 0.01 {
		handlerError(tr(fromID, "lng_priv_give_set_amount_error"))
		return
	}

	// 检查小数点位数
	s := strings.Split(enterAmount, ".")
	if len(s) == 2 && len(s[1]) > 2 {
		handlerError(tr(fromID, "lng_priv_give_set_amount_error"))
		return
	}

	// 检查帐户余额
	balance := getUserAssetAmount(fromID, storage.GetAssetSymbol(info.asset))
	fBalance, _ := strconv.ParseFloat(getUserAssetAmount(fromID, storage.GetAssetSymbol(info.asset)), 10)
	if amount > fBalance {
		reply := tr(fromID, "lng_priv_give_set_amount_no_asset")
		handlerError(fmt.Sprintf(reply, info.asset, balance))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.amount = uint32(amount * 100)
	update.CallbackQuery.Data = data + enterAmount + "/"
	handler.handleLuckyMoneyNumber(bot, r, info, update, false)
}

// 处理红包金额
func (handler *GiveHandler) handleLuckyMoneyAmount(bot *methods.BotExt, r *history.History, info *luckyMoneys,
	update *types.Update) {

	// 处理输入金额
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterLuckyMoneyAmount(bot, r, info, update, back.Message.Text)
		return
	}

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID
	markup := makeGiveBaseMenus(fromID, query.Data)

	// 回复请求结果
	r.Clear().Push(update)
	amount := tr(fromID, "lng_priv_give_amount")
	if info.typ == equalLuckyMoney {
		amount = tr(fromID, "lng_priv_give_value")
	}

	answer := fmt.Sprintf(tr(fromID, "lng_priv_give_set_amount_answer"), amount)
	bot.AnswerCallbackQuery(query, answer, false, "", 0)

	reply := tr(fromID, "lng_priv_give_set_amount")
	bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
	bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)
	reply = fmt.Sprintf(reply, amount, bitCNY, bitUSD, luckyMoneysTypeToString(fromID, info.typ), info.asset)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
}

// 处理输入红包个数
func (handler *GiveHandler) handleEnterLuckyMoneyNumber(bot *methods.BotExt, r *history.History,
	info *luckyMoneys, update *types.Update, enterNumber string) {

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID

	// 处理错误
	handlerError := func(reply string) {
		r.Pop()
		markup := makeGiveBaseMenus(fromID, query.Data)
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 检查红包数量
	number, err := strconv.ParseUint(enterNumber, 10, 32)
	if err != nil {
		handlerError(tr(fromID, "lng_priv_give_set_number_error"))
		return
	}

	// 检查账户余额
	balance := getUserAssetAmount(fromID, storage.GetAssetSymbol(info.asset))
	if info.typ == randLuckyMoney && uint32(number) > info.amount {
		reply := tr(fromID, "lng_priv_give_set_number_no_asset")
		handlerError(fmt.Sprintf(reply, info.asset, balance))
		return
	}

	fBalance, _ := strconv.ParseFloat(balance, 10)
	if info.typ == equalLuckyMoney && (info.amount*uint32(number) > uint32(fBalance*100)) {
		reply := tr(fromID, "lng_priv_give_set_number_no_asset")
		handlerError(fmt.Sprintf(reply, info.asset, balance))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.number = uint32(number)
	update.CallbackQuery.Data += enterNumber + "/"
	handler.handleLuckyMoneyMemo(bot, r, info, update)
}

// 处理红包个数
func (handler *GiveHandler) handleLuckyMoneyNumber(bot *methods.BotExt, r *history.History, info *luckyMoneys,
	update *types.Update, edit bool) {

	// 处理输入个数
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterLuckyMoneyNumber(bot, r, info, update, back.Message.Text)
		return
	}

	// 提示输入红包个数
	r.Clear().Push(update)
	query := update.CallbackQuery
	fromID := query.From.ID
	markup := makeGiveBaseMenus(fromID, query.Data)

	amount := tr(fromID, "lng_priv_give_amount")
	if info.typ == equalLuckyMoney {
		amount = tr(fromID, "lng_priv_give_value")
	}

	reply := ""
	if info.typ == randLuckyMoney {
		reply = tr(fromID, "lng_priv_give_set_number")
		reply = fmt.Sprintf(reply, luckyMoneysTypeToString(fromID, info.typ), info.asset,
			amount, fmt.Sprintf("%.2f", float64(info.amount)/100.0), info.asset)
	} else {
		bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
		bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)
		reply = tr(fromID, "lng_priv_give_set_number_equal")
		reply = fmt.Sprintf(reply, bitCNY, bitUSD, luckyMoneysTypeToString(fromID, info.typ),
			info.asset, amount, fmt.Sprintf("%.2f", float64(info.amount)/100.0), info.asset)
	}

	if !edit {
		bot.SendMessage(fromID, reply, true, markup)
	} else {
		bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
	}
	bot.AnswerCallbackQuery(query, tr(fromID, "lng_priv_give_set_number_answer"), false, "", 0)
}

// 处理输入红包留言
func (handler *GiveHandler) handleEnterLuckyMoneyMemo(bot *methods.BotExt, r *history.History,
	info *luckyMoneys, update *types.Update, memo string) {

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID

	// 处理错误
	handlerError := func(reply string) {
		r.Pop()
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := makeGiveBaseMenus(fromID, query.Data)
		bot.SendMessage(fromID, reply, true, markup)
		return
	}

	// 检查留言长度
	dynamicCfg := config.GetDynamic()
	if len(memo) == 0 || len(memo) > dynamicCfg.MaxMemoLength {
		reply := fmt.Sprintf(tr(fromID, "lng_priv_give_set_memo_error"),
			dynamicCfg.MaxMemoLength)
		handlerError(reply)
		return
	}

	// 处理生成红包
	info.memo = memo
	luckyMoney, err := handler.handleGenerateLuckyMoney(fromID, query.From.FirstName, info)
	if err != nil {
		logger.Warnf("Failed to create lucky money, %v", err)
		handlerError(tr(fromID, "lng_priv_give_create_failed"))
		return
	}

	// 删除已有键盘
	markup := methods.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}

	// 回复红包内容
	r.Clear()
	reply := tr(fromID, "lng_priv_give_created")
	reply = fmt.Sprintf(reply, bot.UserName, luckyMoney.ID, bot.UserName, luckyMoney.ID)
	bot.AnswerCallbackQuery(query, "", false, "", 0)
	bot.SendMessageDisableWebPagePreview(fromID, reply, true, &markup)

	// 回复输入群组ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_chat_enter_chat_id"),
			CallbackData: fmt.Sprintf("/chatid/%d/", luckyMoney.ID),
		},
	}
	reply = tr(fromID, "lng_priv_give_created_enter_chat_id")
	bot.SendMessage(fromID, reply, true, methods.MakeInlineKeyboardMarkupAuto(menus[:], 1))
}

// 处理红包留言
func (handler *GiveHandler) handleLuckyMoneyMemo(bot *methods.BotExt, r *history.History, info *luckyMoneys,
	update *types.Update) {

	// 处理输入留言
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterLuckyMoneyMemo(bot, r, info, update, back.Message.Text)
		return
	}

	// 生成回复键盘
	query := update.CallbackQuery
	fromID := query.From.ID
	menus := [...]methods.KeyboardButton{
		methods.KeyboardButton{
			Text: tr(fromID, "lng_priv_give_benediction"),
		},
	}
	markup := methods.MakeReplyKeyboardMarkup(menus[:], 1)

	// 提示输入红包留言
	r.Clear().Push(update)
	amount := tr(fromID, "lng_priv_give_amount")
	if info.typ == equalLuckyMoney {
		amount = tr(fromID, "lng_priv_give_value")
	}
	reply := tr(fromID, "lng_priv_give_set_memo")
	reply = fmt.Sprintf(reply, luckyMoneysTypeToString(fromID, info.typ), info.asset,
		amount, fmt.Sprintf("%.2f", float64(info.amount)/100.0), info.asset, info.number)
	bot.SendMessage(fromID, reply, true, markup)
	bot.AnswerCallbackQuery(query, tr(fromID, "lng_priv_give_set_memo_answer"), false, "", 0)
}

// 处理生成红包
func (handler *GiveHandler) handleGenerateLuckyMoney(userID int64, firstName string,
	info *luckyMoneys) (*storage.LuckyMoney, error) {

	// 冻结资金
	amount := info.amount
	if info.typ == equalLuckyMoney {
		amount = info.amount * info.number
	}
	assetStorage := storage.AssetStorage{}
	info.asset = storage.GetAssetSymbol(info.asset)
	err := assetStorage.FrozenAsset(userID, info.asset, amount)
	if err != nil {
		return nil, err
	}
	logger.Errorf("Frozen asset, user_id: %v, asset: %v, amount: %v",
		userID, info.asset, amount)

	// 生成红包
	var luckyMoneyArr []int
	if info.typ == randLuckyMoney {
		luckyMoneyArr, err = core.Generate(amount, info.number)
		if err != nil {
			logger.Errorf("Failed to generate lucky money, user_id: %v, %v", userID, err)

			// 解冻资金
			if err = assetStorage.UnfreezeAsset(userID, info.asset, amount); err != nil {
				logger.Errorf("Failed to unfreeze asset, user_id: %v, asset: %v, amount: %v",
					userID, info.asset, amount)
			}
			return nil, err
		}
	} else {
		luckyMoneyArr = make([]int, 0, info.number)
		for i := 0; i < int(info.number); i++ {
			luckyMoneyArr = append(luckyMoneyArr, int(info.amount))
		}
	}

	// 保存红包信息
	luckyMoney := storage.LuckyMoney{
		SenderID:   userID,
		SenderName: firstName,
		Asset:      info.asset,
		Amount:     info.amount,
		Number:     info.number,
		Memo:       info.memo,
		Lucky:      info.typ == randLuckyMoney,
		Timestamp:  time.Now().UTC().Unix(),
	}
	if info.typ == equalLuckyMoney {
		luckyMoney.Value = info.amount
	}
	luckyMoneyStorage := storage.LuckyMoneyStorage{}
	newLuckyMoney, err := luckyMoneyStorage.NewLuckyMoney(&luckyMoney, luckyMoneyArr)
	if err != nil {
		logger.Errorf("Failed to new lucky money, user_id: %v, %v", userID, err)

		// 解冻资金
		if err = assetStorage.UnfreezeAsset(userID, info.asset, amount); err != nil {
			logger.Errorf("Failed to unfreeze asset, user_id: %v, asset: %v, amount: %v",
				userID, info.asset, amount)
		}
		return nil, err
	}
	logger.Errorf("Generate lucky money, id: %v, user_id: %v, asset: %v, amount: %v",
		newLuckyMoney.ID, userID, info.asset, amount)

	// 过期计时
	timer.AddLuckyMoney(luckyMoney.ID, luckyMoney.Timestamp)

	// 记录操作历史
	desc := fmt.Sprintf("您发放了红包(id: *%d*), 花费*%.2f* *%s*", luckyMoney.ID,
		float64(amount)/100.0, luckyMoney.Asset)
	models.InsertHistory(userID, desc)

	return newLuckyMoney, nil
}
