package privatechat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zhangpanyi/botcasino/config"
	"github.com/zhangpanyi/botcasino/httprpc"
	"github.com/zhangpanyi/botcasino/logic/syncfee"
	"github.com/zhangpanyi/botcasino/models"
	"github.com/zhangpanyi/botcasino/storage"
	withdrawservice "github.com/zhangpanyi/botcasino/withdraw"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	tg "github.com/zhangpanyi/basebot/telegram"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
)

// 匹配资产
var reMathWithdrawAsset *regexp.Regexp

// 匹配金额
var reMathWithdrawAmount *regexp.Regexp

// 匹配账户
var reMathWithdrawAccout *regexp.Regexp

// 匹配提交
var reMathWithdrawSubmit *regexp.Regexp

func init() {
	var err error
	reMathWithdrawAsset, err = regexp.Compile("^/withdraw/(\\w+)/$")
	if err != nil {
		panic(err)
	}

	reMathWithdrawAmount, err = regexp.Compile("^/withdraw/(\\w+)/([0-9]+\\.?[0-9]*)/$")
	if err != nil {
		panic(err)
	}

	reMathWithdrawAccout, err = regexp.Compile("^/withdraw/(\\w+)/([0-9]+\\.?[0-9]*)/(\\w+)/$")
	if err != nil {
		panic(err)
	}

	reMathWithdrawSubmit, err = regexp.Compile("^/withdraw/(\\w+)/([0-9]+\\.?[0-9]*)/([\\w|-]+)/submit/$")
	if err != nil {
		panic(err)
	}
}

// WithdrawHandler 取款
type WithdrawHandler struct {
}

// 取款信息
type withdrawInfo struct {
	account string // 账户名
	asset   string // 资产类型
	amount  uint32 // 资产数量
}

// Handle 消息处理
func (handler *WithdrawHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 处理选择资产
	data := update.CallbackQuery.Data
	if data == "/withdraw/" {
		r.Clear()
		dynamicCfg := config.GetDynamic()
		if dynamicCfg.AllowWithdraw {
			handler.handleChooseAsset(bot, update.CallbackQuery)
		} else {
			// 未开放提现
			fromID := update.CallbackQuery.From.ID
			reply := tr(fromID, "lng_priv_withdraw_not_allow")
			bot.AnswerCallbackQuery(update.CallbackQuery, reply, false, "", 0)
		}
		return
	}

	// 处理输入金额
	var info withdrawInfo
	result := reMathWithdrawAsset.FindStringSubmatch(data)
	if len(result) == 2 {
		info.asset = result[1]
		handler.handleWithdrawAmount(bot, r, &info, update)
		return
	}

	// 处理输入账户名
	result = reMathWithdrawAmount.FindStringSubmatch(data)
	if len(result) == 3 {
		info.asset = result[1]
		amount, _ := strconv.ParseFloat(result[2], 10)
		info.amount = uint32(amount * 100)
		handler.handleWithdrawAccout(bot, r, &info, update, true)
		return
	}

	// 处理提现总览
	result = reMathWithdrawAccout.FindStringSubmatch(data)
	if len(result) == 4 {
		info.asset = result[1]
		amount, _ := strconv.ParseFloat(result[2], 10)
		info.amount = uint32(amount * 100)
		info.account = result[3]
		handler.handleWithdrawOverview(bot, r, &info, update, true)
		return
	}

	// 处理提现请求
	result = reMathWithdrawSubmit.FindStringSubmatch(data)
	if len(result) == 4 {
		dynamicCfg := config.GetDynamic()
		if dynamicCfg.AllowWithdraw {
			info.asset = result[1]
			amount, _ := strconv.ParseFloat(result[2], 10)
			info.amount = uint32(amount * 100)
			info.account = result[3]
			handler.handleWithdraw(bot, r, &info, update.CallbackQuery)
		} else {
			// 未开放提现
			fromID := update.CallbackQuery.From.ID
			reply := tr(fromID, "lng_priv_withdraw_not_allow")
			bot.AnswerCallbackQuery(update.CallbackQuery, reply, false, "", 0)

			menus := [...]methods.InlineKeyboardButton{
				methods.InlineKeyboardButton{Text: tr(fromID, "lng_back_superior"), CallbackData: "/main/"},
			}
			markup := methods.MakeInlineKeyboardMarkup(menus[:], 1)
			bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, false, markup)
		}
		return
	}

	// 无效帐户名处理
	fromID := update.CallbackQuery.From.ID
	reply := tr(fromID, "lng_priv_withdraw_account_error")
	bot.AnswerCallbackQuery(update.CallbackQuery, reply, true, "", 0)
}

// 消息路由
func (handler *WithdrawHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}

// 处理选择资产
func (handler *WithdrawHandler) handleChooseAsset(bot *methods.BotExt, query *types.CallbackQuery) {
	// 生成菜单列表
	data := query.Data
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{Text: "💴 bitCNY", CallbackData: data + storage.BitCNY + "/"},
		methods.InlineKeyboardButton{Text: "💵 bitUSD", CallbackData: data + storage.BitUSD + "/"},
		methods.InlineKeyboardButton{Text: tr(fromID, "lng_back_superior"), CallbackData: "/main/"},
	}

	// 获取资产信息
	bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
	bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)

	// 回复请求结果
	bot.AnswerCallbackQuery(query, "", false, "", 0)
	markup := methods.MakeInlineKeyboardMarkup(menus[:], 2, 1, 1)
	reply := fmt.Sprintf(tr(fromID, "lng_priv_withdraw_say"), bitCNY, bitUSD)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
}

// 处理输入提现金额
func (handler *WithdrawHandler) handleEnterWithdrawAmount(bot *methods.BotExt, r *history.History,
	info *withdrawInfo, update *types.Update, amount string) {

	// 处理错误
	query := update.CallbackQuery
	fromID := query.From.ID
	data := query.Data
	handlerError := func(reply string) {
		r.Pop()
		menus := [...]methods.InlineKeyboardButton{
			methods.InlineKeyboardButton{
				Text:         tr(fromID, "lng_back_superior"),
				CallbackData: backSuperior(data),
			},
		}
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 获取资产信息
	bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
	bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)

	// 检查输入金额
	result := strings.Split(amount, ".")
	if len(result) == 2 && len(result[1]) > 2 {
		fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
		reply := tr(fromID, "lng_priv_withdraw_amount_not_enough")
		handlerError(fmt.Sprintf(reply, info.asset, bitCNY, bitUSD,
			fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset))
		return
	}

	fAmount, err := strconv.ParseFloat(amount, 10)
	if err != nil {
		fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
		reply := tr(fromID, "lng_priv_withdraw_amount_not_enough")
		handlerError(fmt.Sprintf(reply, info.asset, bitCNY, bitUSD,
			fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset))
		return
	}

	// 检查用户余额
	lAmount := uint32(fAmount * 100)
	newHandler := storage.AssetStorage{}
	fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
	asset, err := newHandler.GetAsset(fromID, storage.GetAssetSymbol(info.asset))
	if err != nil || asset.Amount < (lAmount+fee) {
		reply := tr(fromID, "lng_priv_withdraw_amount_error")
		handlerError(fmt.Sprintf(reply, info.asset, bitCNY, bitUSD,
			fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.amount = lAmount
	update.CallbackQuery.Data = data + amount + "/"
	handler.handleWithdrawAccout(bot, r, info, update, false)
}

// 处理提现金额
func (handler *WithdrawHandler) handleWithdrawAmount(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update) {

	// 处理输入个数
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterWithdrawAmount(bot, r, info, update, back.Message.Text)
		return
	}

	// 提示输入提现金额
	r.Clear().Push(update)
	query := update.CallbackQuery
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(query.Data),
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
	bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
	bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)
	reply := tr(fromID, "lng_priv_withdraw_enter_amount")
	reply = fmt.Sprintf(reply, info.asset, info.asset, bitCNY, bitUSD,
		fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)

	answer := tr(fromID, "lng_priv_withdraw_enter_amount_answer")
	answer = fmt.Sprintf(answer, info.asset)
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
}

// 处理输入账户名
func (handler *WithdrawHandler) handleEnterWithdrawAccout(bot *methods.BotExt, r *history.History,
	info *withdrawInfo, update *types.Update, account string) {

	// 处理错误
	query := update.CallbackQuery
	fromID := query.From.ID
	data := query.Data
	handlerError := func(reply string) {
		r.Pop()
		menus := [...]methods.InlineKeyboardButton{
			methods.InlineKeyboardButton{
				Text:         tr(fromID, "lng_back_superior"),
				CallbackData: backSuperior(data),
			},
		}
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 检查帐号长度
	if len(account) == 0 || len(account) > 32 {
		handlerError(tr(fromID, "lng_priv_withdraw_account_error"))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.account = account
	update.CallbackQuery.Data = data + account + "/"
	handler.handleWithdrawOverview(bot, r, info, update, false)
}

// 处理提现账户名
func (handler *WithdrawHandler) handleWithdrawAccout(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update, edit bool) {

	// 处理输入金额
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterWithdrawAccout(bot, r, info, update, back.Message.Text)
		return
	}

	// 获取资产信息
	query := update.CallbackQuery
	fromID := query.From.ID

	// 生成菜单列表
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(query.Data),
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 回复请求结果
	r.Clear().Push(update)
	reply := tr(fromID, "lng_priv_withdraw_enter_account")
	reply = fmt.Sprintf(reply, fmt.Sprintf("%.2f", float64(info.amount)/100.0), info.asset)
	if !edit {
		bot.SendMessage(fromID, reply, true, markup)
	} else {
		bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
	}

	answer := tr(fromID, "lng_priv_withdraw_enter_account_answer")
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
}

// 处理提现概览
func (handler *WithdrawHandler) handleWithdrawOverview(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update, edit bool) {

	fromID := update.CallbackQuery.From.ID
	answer := tr(fromID, "lng_priv_withdraw_overview_answer")
	bot.AnswerCallbackQuery(update.CallbackQuery, answer, false, "", 0)

	fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
	sfee := fmt.Sprintf("%.2f", float64(fee)/100.0)
	reply := tr(fromID, "lng_priv_withdraw_overview")
	amount := fmt.Sprintf("%.2f", float64(info.amount)/100.0)
	reply = fmt.Sprintf(reply, info.account, info.asset, amount, info.asset,
		amount, sfee, info.asset, sfee, info.asset)

	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_priv_withdraw_submit"),
			CallbackData: update.CallbackQuery.Data + "submit/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(update.CallbackQuery.Data),
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
	if !edit {
		bot.SendMessage(fromID, reply, true, markup)
	} else {
		bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, true, markup)
	}
}

// 处理提现
func (handler *WithdrawHandler) handleWithdraw(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	query *types.CallbackQuery) {

	// 生成菜单
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_menu"),
			CallbackData: "/main/",
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 获取手续费
	asset := storage.GetAssetSymbol(info.asset)
	fee, _ := syncfee.GetFee(asset)

	// 扣除余额
	newHandler := storage.AssetStorage{}
	err := newHandler.Withdraw(fromID, asset, info.amount+fee)
	if err != nil {
		logger.Warnf("Failed to withdraw asset, UserID: %d, Asset: %s, Amount: %d, Fee: %d, %v",
			fromID, asset, info.amount, fee, err)
		reply := tr(fromID, "lng_priv_withdraw_no_money")
		bot.AnswerCallbackQuery(query, reply, false, "", 0)
		bot.EditMessageReplyMarkup(query.Message, reply, false, markup)
		return
	}
	logger.Errorf("Withdraw asset success, UserID: %d, Asset: %s, Amount: %d, Fee: %d",
		fromID, asset, info.amount, fee)

	// 提交成功
	reply := tr(fromID, "lng_priv_withdraw_submit_ok")
	answer := tr(fromID, "lng_priv_withdraw_submit_ok_answer")
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
	bot.EditMessageReplyMarkup(query.Message, reply, true, nil)

	// 钱包转账
	assetID := httprpc.USDAssetID
	if asset != storage.BitUSDSymbol {
		assetID = httprpc.CNYAssetID
	}
	future, err := withdrawservice.AddFuture(fromID, info.account, assetID, info.amount, fee)

	// 记录操作历史
	desc := fmt.Sprintf("您申请提现*%.2f* *%s*到比特股账户 *%s* 正在处理(订单ID: *%d*), 花费手续费*%.2f* *%s*",
		float64(info.amount)/100.0, asset, tg.Pre(info.account), future.OrderID, float64(fee)/100.0, asset)
	models.InsertHistory(fromID, desc)

	// 获取转账结果
	err = handler.HandleWithdrawFuture(future)
	if err != nil {
		// 返回处理结果
		reply := tr(fromID, "lng_priv_withdraw_wallet_error")
		bot.EditMessageReplyMarkup(query.Message, reply, false, markup)
		return
	}
	logger.Errorf("Transfer asset success, OrderID: %d", future.OrderID)

	// 返回结果
	reply = tr(fromID, "lng_priv_withdraw_success")
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
	return
}

// 处理提现结果
func (handler *WithdrawHandler) HandleWithdrawFuture(future *withdrawservice.Future) error {
	err := future.GetResult()
	if err == nil {
		return nil
	}
	logger.Errorf("Failed to withdraw asset, transfer error, OrderID: %d, %v", future.OrderID, err)

	asset := storage.BitCNYSymbol
	if future.Transfer.AssetID != httprpc.CNYAssetID {
		asset = storage.BitUSDSymbol
	}

	// 退还资金
	newHandler := storage.AssetStorage{}
	amount := future.Transfer.Amount + future.Transfer.Fee
	if err2 := newHandler.Deposit(future.Transfer.UserID, asset, amount); err2 != nil {
		logger.Errorf("Failed to return withdraw asset, OrderID: %d, %v", future.OrderID, err2)

		// 记录操作历史
		desc := fmt.Sprintf("您申请提现*%.2f* *%s*到比特股账户 *%s* 处理失败(订单ID: *%d*), 退还资金失败",
			float64(future.Transfer.Amount)/100.0, asset, tg.Pre(future.Transfer.To), future.OrderID)
		models.InsertHistory(future.Transfer.UserID, desc)
		return err2
	}

	// 记录操作历史
	desc := fmt.Sprintf("您申请提现*%.2f* *%s*到比特股账户 *%s* 处理失败(订单ID: *%d*), 已退还资金",
		float64(future.Transfer.Amount)/100.0, asset, tg.Pre(future.Transfer.To), future.OrderID)
	models.InsertHistory(future.Transfer.UserID, desc)
	return err
}
