package bots

import (
	"github.com/forchain/cryptotrader/binance"
	"context"
	"time"
	"github.com/sirupsen/logrus"
	"github.com/forchain/cryptotrader/model"
	"io/ioutil"
	"github.com/tidwall/gjson"
	"fmt"
	"github.com/forchain/TradeBot/helpers"
)

type BinanceBot struct {
	Bot

	client_ *binance.Client
}

func (_b *BinanceBot) Init() {
	_b.Bot.Init("binance", 0.9995)

	_b.client_ = binance.New("o6GLpg2HICo0ziUoyXhRgjriE0Pl717cXcp3CGEISftxZ01xOnDEviHmkCqMU58Q", "Km4uGLJHPiBxdEEA1JblVMeUqlYa4wtnvjKEyPLA0BNvbQphsRrgKuocCY0uldN5")
}


func (_b *BinanceBot) GetPairs() map[string][]string {
	filename := fmt.Sprintf("%v/data/%v/exchangeInfo.json", helpers.WorkingDir(), _b.Name)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Fatalln(err)
	}

	coinMap := make(map[string][]string)
	jsonCoins := gjson.GetBytes(b, `symbols`).Array()
	for _, c := range jsonCoins {
		base := c.Get("baseAsset").Str
		quote := c.Get("quoteAsset").Str
		coinMap[quote] = append(coinMap[quote], base)
	}

	return coinMap
}

func (_b *BinanceBot) GetRecords(_base, _quote string) (records []model.Record, err error) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	records, err = _b.client_.GetRecords(context, _base, _quote, "1h", 0, 0, int64(_b.Bot.Limit))
	defer cancel()

	return
}
