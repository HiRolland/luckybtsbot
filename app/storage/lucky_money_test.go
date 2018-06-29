package storage_test

import (
	"log"
	"testing"

	"github.com/zhangpanyi/botcasino/app/logic/core"
	"github.com/zhangpanyi/botcasino/app/storage"
)

// 测试创建红包
func TestNewLuckyMoney(t *testing.T) {
	storage.Connect("test.db")

	number := 1
	arr, err := core.Generate(10000, uint32(number))
	if err != nil {
		t.Fatal(err)
	}

	luckyMoney := &storage.LuckyMoney{
		Asset:  "bitCNY",
		Amount: 100,
		Number: uint32(number),
	}
	handler := storage.LuckyMoneyStorage{}
	luckyMoney, err = handler.NewLuckyMoney(luckyMoney, arr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(luckyMoney)
}

// 测试获取红包信息
func TestGetLuckyMoney(t *testing.T) {
	storage.Connect("test.db")

	handler := storage.LuckyMoneyStorage{}
	luckyMoney, received, err := handler.GetLuckyMoney(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(luckyMoney, received)
}

// 测试领取红包
func TestReceiveLuckyMoney(t *testing.T) {
	storage.Connect("test.db")

	handler := storage.LuckyMoneyStorage{}
	for i := 0; i < 100; i++ {
		amount, number, err := handler.ReceiveLuckyMoney(1, int64(i), "zpy")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(amount, number)
	}
}

// 测试获取极端红包
func TestGetTwoTxtremes(t *testing.T) {
	storage.Connect("test.db")

	handler := storage.LuckyMoneyStorage{}
	min, max, err := handler.GetTwoTxtremes(100002)
	if err != nil {
		log.Fatalln(err)
	}
	t.Log(min, max)
}

// 测试遍历红包
func TestForeachLuckyMoney(t *testing.T) {
	storage.Connect("test.db")

	handler := storage.LuckyMoneyStorage{}
	err := handler.ForeachLuckyMoney(100043, func(data *storage.LuckyMoney) {
		log.Println(data)
	})
	if err != nil {
		log.Fatalln(err)
	}
}
