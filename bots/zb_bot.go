package bots

import (
	"github.com/forchain/cryptotrader/zb"
	"os"
	"github.com/forchain/cryptotrader/model"
	"fmt"
	"io/ioutil"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/forchain/TradeBot/helpers"
	"strings"
	//"encoding/json"
	"time"
)

type ZbBot struct {
	Bot

	client_ *zb.ZB
}

func (_b *ZbBot) Init() {
	_b.Bot.Init("zb", 0.998)

	_b.client_ = zb.New(os.Getenv("ZB_ACCESS_KEY"), os.Getenv("ZB_SECRET_KEY"))
}

func (_b *ZbBot) GetRecords(_base, _quote string) (records []model.Record, err error) {
	records, err = _b.client_.GetRecords(_base, _quote, "1hour", 0, _b.Limit)
	time.Sleep(time.Millisecond * 1500)
	return
}

func (_b *ZbBot) GetPairs() map[string][]string {
	filename := fmt.Sprintf("%v/data/%v/markets.json", helpers.WorkingDir(), _b.Name)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Fatalln(err)
	}

	coinMap := make(map[string][]string)
	jsonCoins := gjson.GetBytes(b, "0").Map()
	for key := range jsonCoins {
		tokens := strings.Split(key, "_")
		base := tokens[0]
		quote := tokens[1]
		coinMap[quote] = append(coinMap[quote], base)
	}

	return coinMap
}
