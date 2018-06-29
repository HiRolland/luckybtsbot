package config

import (
	"sync"

	"gopkg.in/yaml.v2"
)

// 动态配置
type Dynamic struct {
	DynamicCfg
	lock sync.RWMutex
}

// 手续费配置
type FeeCfg struct {
	CNY uint32 `yaml:"cny"` // 人民币手续费
	USD uint32 `yaml:"usd"` // 美元手续费
}

// 配置数据
type DynamicCfg struct {
	WhiteList        []string `yaml:"white_list"`         // 地址白名单
	Suspended        bool     `yaml:"suspended"`          // 是否暂停服务
	MaxMemoLength    int      `yaml:"max_memo_length"`    // 备注最大长度
	AllowDeposit     bool     `yaml:"allow_deposit"`      // 是否允许充值
	AllowWithdraw    bool     `yaml:"allow_withdraw"`     // 是否允许提现
	LuckyMoneyExpire int64    `yaml:"lucky_money_expire"` // 红包过期时间
	UseFixedFee      bool     `yaml:"use_fixed_fee"`      // 是否使用固定手续费
	FixedFeeAmount   FeeCfg   `yaml:"fixed_fee_amount"`   // 固定手续费
}

// 解析数据
func (d *Dynamic) parse(data []byte) error {
	cfg := DynamicCfg{}
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	d.DynamicCfg = cfg
	return nil
}
