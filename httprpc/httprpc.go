package httprpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/zhangpanyi/botcasino/config"
)

// JSON-RPC错误
type Error struct {
	Code    int32  `json:"code"`    // 代码
	Message string `json:"message"` // 详情
}

// JSON-RPC请求
type Request struct {
	Version string      `json:"jsonrpc"` // 版本
	ID      int64       `json:"id"`      // 请求ID
	Method  string      `json:"method"`  // 方法
	Params  interface{} `json:"params"`  // 参数
}

// JSON_RPC相应
type Response struct {
	Version string          `json:"jsonrpc"` // 版本
	ID      int64           `json:"id"`      // 请求ID
	Error   *Error          `json:"error"`   // 错误信息
	Result  json.RawMessage `json:"result"`  // 返回结果
}

// 方法调用
func (req *Request) Call() ([]byte, error) {
	jsb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	serveCfg := config.GetServe()
	resp, err := http.Post(serveCfg.MonitorURL, "application/json", bytes.NewReader(jsb))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res Response
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, errors.New(res.Error.Message)
	}
	return res.Result, nil
}

// 生成请求
func MakeRequest(method string, params ...interface{}) *Request {
	req := new(Request)
	req.ID = 1
	req.Version = "2.0"
	req.Method = method
	req.Params = params
	return req
}

// GetFees 获取手续费
func GetFees(assets []string) ([]uint32, error) {
	request := MakeRequest("get_transfer_fees", assets)
	jsb, err := request.Call()
	if err != nil {
		return nil, err
	}

	fees := make([]uint32, 0)
	if err = json.Unmarshal(jsb, fees); err != nil {
		return nil, err
	}
	return fees, nil
}

// Transfer 转账操作
func Transfer(orderID int64, to, assetID string, amount uint32) error {
	memo := strconv.FormatInt(orderID, 10)
	request := MakeRequest("transfer", to, assetID, amount, memo)
	_, err := request.Call()
	return err
}
