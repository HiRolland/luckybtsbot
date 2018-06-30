package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/botcasino/app/config"
	"github.com/zhangpanyi/botcasino/app/logic/notice"
	"github.com/zhangpanyi/botcasino/app/models"
	"github.com/zhangpanyi/botcasino/app/storage"
)

// 匹配用户ID
var reMathUserID *regexp.Regexp

func init() {
	var err error
	reMathUserID, err = regexp.Compile("^ *(\\d+) *$")
	if err != nil {
		panic(err)
	}
}

// 手续费
type Fee struct {
	Asset   string  `json:"asset"`    // 资产符号
	AssetID string  `json:"asset_id"` // 资产ID
	Amount  float64 `json:"amount"`   // 金额
}

// 转账操作
type TransferOperation struct {
	ID        string  `json:"trx_id"`    // 操作ID
	BlockNum  int64   `json:"block_num"` // 区块高度
	Asset     string  `json:"asset"`     // 资产符号
	AssetID   string  `json:"asset_id"`  // 资产ID
	Amount    float64 `json:"amount"`    // 金额
	Fee       *Fee    `json:"fee"`       // 手续费
	From      string  `json:"from"`      // 来源
	To        string  `json:"to"`        // 目标
	Memo      *string `json:"memo"`      // 备注
	Timestamp string  `json:"timestamp"` // 时间戳
}

// 转账操作回调
func handleTransferOperation(w http.ResponseWriter, r *http.Request) {
	// 读取数据
	jsb, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.Infof("Webhook, failed to read body, %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 序列化请求
	var operation TransferOperation
	if err = json.Unmarshal(jsb, &operation); err != nil {
		logger.Infof("Webhook, failed to parse transfer operation, %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	serveCfg := config.GetServe()
	logger.Infof("Webhook, new transfer operation: %+v, %v", operation, serveCfg.Account)

	// 请求分发处理
	var memo string
	if operation.Memo != nil {
		memo = *operation.Memo
	}
	if operation.To == serveCfg.Account {
		handleTransferredIn(operation.Asset, operation.Amount, operation.BlockNum, memo)
	} else if operation.From == serveCfg.Account {
		handleTransferredOut(operation.Asset, operation.Amount, operation.Fee.Amount, operation.BlockNum, memo)
	}

	w.WriteHeader(http.StatusOK)
}

// 处理转入操作
func handleTransferredIn(asset string, amount float64, blockNum int64, memo string) {
	logger.Warnf("On receive notice, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s",
		asset, amount, blockNum, memo)

	// 获取用户ID
	result := reMathUserID.FindStringSubmatch(memo)
	if len(result) != 2 {
		logger.Warnf("On receive notice, not found user id, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s",
			asset, amount, blockNum, memo)
		return
	}

	userID, err := strconv.ParseInt(result[1], 10, 64)
	if err != nil {
		logger.Warnf("On receive notice, not found user id, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s",
			asset, amount, blockNum, memo)
		return
	}

	// 检查资产类型
	if asset != storage.BitCNYSymbol && asset != storage.BitUSDSymbol {
		logger.Warnf("On receive notice, nonsupport asset, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s",
			asset, amount, blockNum, memo)
		return
	}

	// 增加用户资产
	realAmount := uint32(amount * 100)
	handler := storage.AssetStorage{}
	err = handler.Deposit(userID, asset, realAmount)
	if err != nil {
		logger.Errorf("On receive notice, deposit failure, UserID=%d, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s, %v",
			userID, asset, amount, blockNum, memo, err)
		return
	}
	logger.Errorf("On receive notice, deposit success, UserID=%d, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s",
		userID, asset, amount, blockNum, memo)

	// 插入操作记录
	desc := fmt.Sprintf("您充值*%.2f* *%s*已确认, 区块高度: *%d*", amount, asset, blockNum)
	models.InsertHistory(userID, desc)

	// 推送充值通知
	notice.SendNotice(userID, desc)
}

// 处理转出操作
func handleTransferredOut(asset string, amount, fee float64, blockNum int64, memo string) {
	logger.Warnf("On sent notice, Asset=%s, Amount=%f, BlockNum=%d, Memo=%s",
		asset, amount, blockNum, memo)

	// 获取订单ID
	orderID, err := strconv.ParseInt(memo, 10, 64)
	if err != nil {
		logger.Warnf("On sent notice, invalid order id, %s", memo)
		return
	}

	// 获取订单信息
	order, err := models.GetWithdrawOrder(orderID)
	if err != nil {
		logger.Warnf("On sent notice, nou found order id, %d", orderID)
		return
	}

	// 更新手续费
	models.UpdateWithdrawFee(orderID, uint32(fee*100))

	// 插入操作记录
	desc := fmt.Sprintf("您提现*%.2f* *%s*已确认, 区块高度: *%d*", amount, asset, blockNum)
	models.InsertHistory(order.UserID, desc)
}
