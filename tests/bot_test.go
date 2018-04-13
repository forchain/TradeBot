package tests

import (
	"testing"
	"github.com/forchain/TradeBot/bots"
	"github.com/sirupsen/logrus"
	"strings"
)

func TestBinance_SortCoins(_t *testing.T) {
	bot := new(bots.BinanceBot)
	bot.Init()
	bots.SortCoins(bot, "", "")
}

func TestBinance_GenerateReport(_t *testing.T) {
	bot := new(bots.BinanceBot)
	bot.Init()
	//bot.Limit = 200

	bridge := ""
	base := "HSR"
	quote := "BTC"

	bridge = strings.ToUpper(bridge)
	base = strings.ToUpper(base)
	quote = strings.ToUpper(quote)

	logrus.SetLevel(logrus.DebugLevel)
	data := bots.LoadData(bot)
	if bridge == "" {
		records := data[quote][base]
		bots.GenerateReport(bot, records, nil)
	} else {
		br2b := data[bridge][base]
		b2q := data[quote][bridge]
		bots.GenerateReport(bot, br2b, b2q)
	}
}

func TestBinance_SaveData(_t *testing.T) {
	bot := new(bots.BinanceBot)
	bot.Init()
	bots.SaveData(bot)
}

func TestBinance_LoadData(_t *testing.T) {
	bot := new(bots.BinanceBot)
	bot.Init()
	coinMap := bots.LoadData(bot)

	logrus.Info(len(coinMap))
}

func TestBinance_GetPairs(_t *testing.T) {
	bot := new(bots.BinanceBot)
	bot.Init()
	pairs := bot.GetPairs()

	logrus.Info(len(pairs))
}

func TestZb_SaveData(_t *testing.T) {
	bot := new(bots.ZbBot)
	bot.Init()
	bots.SaveData(bot)
}

func TestZb_GetPairs(_t *testing.T) {
	bot := new(bots.ZbBot)
	bot.Init()
	pairs := bot.GetPairs()

	logrus.Info(len(pairs))
}

func TestZb_LoadData(_t *testing.T) {
	bot := new(bots.ZbBot)
	bot.Init()
	coinMap := bots.LoadData(bot)

	logrus.Info(len(coinMap))
}

func TestZb_SortCoins(_t *testing.T) {
	bot := new(bots.ZbBot)
	bot.Init()
	bots.SortCoins(bot, "", "")
}

func TestZb_GenerateReport(_t *testing.T) {
	bot := new(bots.ZbBot)
	bot.Init()
	//bot.Limit = 200

	base := "btn"
	quote := "btc"
	bridge := ""

	logrus.SetLevel(logrus.DebugLevel)
	data := bots.LoadData(bot)
	if bridge == "" {
		records := data[quote][base]
		bots.GenerateReport(bot, records, nil)
	} else {
		br2b := data[bridge][base]
		b2q := data[quote][bridge]
		bots.GenerateReport(bot, br2b, b2q)
	}
}
