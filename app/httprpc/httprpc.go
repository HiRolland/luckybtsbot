package httprpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/zhangpanyi/botcasino/app/config"
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

func toString(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}
	str := string(raw)
	if len(str) > 2 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return string(raw)
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
	if len(params) == 0 {
		params = make([]interface{}, 0)
	}
	req := new(Request)
	req.ID = 1
	req.Version = "2.0"
	req.Method = method
	req.Params = params
	return req
}

// 获取系统账户
func GetAccount() (string, error) {
	request := MakeRequest("account")
	jsb, err := request.Call()
	if err != nil {
		return "", err
	}
	return toString(jsb), nil
}

// 获取手续费
func GetFees(assets []string) ([]uint32, error) {
	request := MakeRequest("get_transfer_fees", assets)
	jsb, err := request.Call()
	if err != nil {
		return nil, err
	}

	fees := make([]float64, 0, len(assets))
	if err = json.Unmarshal(jsb, &fees); err != nil {
		return nil, err
	}

	result := make([]uint32, 0, len(assets))
	for i := 0; i < len(fees); i++ {
		result = append(result, uint32(fees[i]*100))
	}
	return result, nil
}

// 转账操作
func Transfer(orderID int64, to, asset string, amount uint32) error {
	memo := strconv.FormatInt(orderID, 10)
	request := MakeRequest("transfer", to, asset, float64(amount)/100.0, memo)
	_, err := request.Call()
	return err
}
