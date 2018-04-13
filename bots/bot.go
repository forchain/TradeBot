package bots

import (
	"github.com/forchain/cryptotrader/model"
	"github.com/markcheno/go-talib"
	"github.com/sirupsen/logrus"
	"math"
	"sort"
	"os"
	"encoding/gob"
	"github.com/forchain/TradeBot/helpers"
	"fmt"
)

type IBot interface {
	Init()
	GetRecords(_base string, _quote string) ([]model.Record, error)
	GetPairs() map[string][]string
	GetFee() float64
	GetName() string
}

type Bot struct {
	IBot
	Fee   float64
	Limit int
	Name  string
}

func (_b *Bot) GetFee() float64 {
	return _b.Fee
}

func (_b *Bot) GetName() string {
	return _b.Name
}

func (_b *Bot) Init(_name string, _fee float64) {
	_b.Fee = _fee
	_b.Name = _name
}

func ConvertRecords(_base, _quote []model.Record) []model.Record {
	result := make([]model.Record, 0)

	baseSize := len(_base)
	quoteSize := len(_quote)

	if baseSize > quoteSize {
		logrus.Fatalln("baseSize:%v > quoteSize:%v ", baseSize, quoteSize)
	} else if baseSize < quoteSize {
		_quote = _quote[quoteSize-baseSize:]

		if len(_quote) != len(_base) {
			logrus.Fatalf("len(to):%v != len(from):%v", len(_quote), len(_base))
		}
	}

	for i, b := range _base {
		//if _base[i].Time != _quote[i].Time {
		//	logrus.Infof("k:%v from:%v != to:%v, err: %v", i, _base[i].Time, _quote[i].Time, "different time")
		//}

		r := model.Record{
			Open:  b.Open * _quote[i].Open,
			Close: b.Close * _quote[i].Close,
			Time:  b.Time,
			Vol:   b.Vol * (_quote[i].Open + _quote[i].Close + _quote[i].High + _quote[i].Low) / 4,
		}

		high, low := r.Close, r.Open
		if high < low {
			high, low = low, high
		}

		r.High = math.Sqrt(r.High * b.High * high)
		r.Low = math.Sqrt(r.Low * b.Low * low)
		result = append(result, r)
	}

	return result
}

type tScore struct {
	Base  string
	Quote string
	Score float64
}

type tByScore []tScore

func (a tByScore) Len() int           { return len(a) }
func (a tByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a tByScore) Less(i, j int) bool { return a[i].Score > a[j].Score }

// for example, _quote: usdt, _bridge: bnb
func SaveData(_b IBot) {
	coinMap := make(map[string]map[string][]model.Record)

	pairs := _b.GetPairs()

	for quote, bases := range pairs {
		for k, base := range bases {
			if records, err := _b.GetRecords(base, quote); err == nil {
				baseMap, ok := coinMap[quote]
				if !ok {
					baseMap = make(map[string][]model.Record)
				}
				baseMap[base] = records
				coinMap[quote] = baseMap
				logrus.Infof("k:%v base:%v quote:%v records:%v", k, base, quote, len(records))
			} else {
				logrus.Fatalln(err)
			}
		}
	}

	workingDir := helpers.WorkingDir()
	filePath := fmt.Sprintf("%v/data/%v/kline.gob", workingDir, _b.GetName())
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(coinMap)
	}
	file.Close()

}

func LoadData(_b IBot) map[string]map[string][]model.Record {
	filePath := fmt.Sprintf("%v/data/%v/kline.gob", helpers.WorkingDir(), _b.GetName())
	file, err := os.Open(filePath)

	coinMap := make(map[string]map[string][]model.Record)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&coinMap)
		if err != nil {
			logrus.Fatalln(err)
		}
	}
	file.Close()

	return coinMap
}

func SortCoins(_b IBot, _quote string, _bridge string) {

	rank := make(tByScore, 0)

	coinMap := LoadData(_b)
	if _quote == "" {
		if _bridge == "" {
			for quote, bases := range coinMap {
				for base, records := range bases {
					score := GenerateReport(_b, records, nil)
					rank = append(rank, tScore{base, quote, score})
				}
			}
		} else {
			logrus.Error("empty bridge with quote")
		}

	} else {
		if _bridge == "" {
			bases := coinMap[_quote]
			for base, records := range bases {
				if  len(records) < 100 {
					continue
				}
				score := GenerateReport(_b, records, nil)
				rank = append(rank, tScore{base, _quote, score})
			}
		} else {
			bridge := coinMap[_quote][_bridge]
			bases := coinMap[_bridge]
			for base, records := range bases {
				logrus.Infof("base:%v _quote:%v _bridge:%v", base, _quote, _bridge)
				score := GenerateReport(_b, records, bridge)
				rank = append(rank, tScore{base, _quote, score})
			}
		}
	}

	sort.Sort(rank)
	for _, v := range rank {
		logrus.Info(v.Base, v.Score)
	}
}

func GenerateReport(_b IBot, _records []model.Record, _bridge []model.Record) float64 {
	records := _records
	if _bridge != nil && len(_bridge) >= len(records) {
		records = ConvertRecords(records, _bridge)
	}

	inReal := make([]float64, len(records))
	for k, v := range records {
		inReal[k] = v.Close
	}
	outMACD, _, outMACDHist := talib.MacdFix(inReal, 9)

	startBase := 100.0
	startQuote := 0.0
	startAsset := startQuote + startBase*records[32].Close

	base := startBase
	quote := startQuote

	// >0 gold < 0 dead
	macd := make([]int, 0)

	buyPrice := -1.0
	sellPrice := -1.0
	buyNum := 0
	sellNum := 0

	// 0~31 are empty

	// -1 means excludes last
	for i := 32; i < len(outMACD)-1; i++ {

		latest := _records[i]
		marketPrice := _records[i+1].Open

		avgPrice := (latest.Open + latest.Close + latest.High + latest.Low) / 4

		timeStr := records[i].Time.Format("01-02 15")
		doBuyOrder := func() {
			buyPrice = latest.Low
			if latest.Open < latest.Close {
				buyPrice = avgPrice
			}
			if buyPrice > marketPrice {
				buyPrice = marketPrice
			}
			orderBase := quote / buyPrice
			logrus.Debugf("%v[%v][BUY ORDER]quote:%v buyPrice:%v orderBase:%v", i, timeStr, quote, buyPrice, orderBase)
		}

		doSellOrder := func() {
			sellPrice = latest.High
			if latest.Open > latest.Close {
				sellPrice = avgPrice
			}
			if sellPrice < marketPrice {
				sellPrice = marketPrice
			}
			orderQuote := sellPrice * base

			logrus.Debugf("%v[%v][SELL ORDER]base:%v sellPrice:%v orderQuote:%v", i, timeStr, base, sellPrice, orderQuote)
		}
		if buyPrice > 0 {
			// 失败, 定价太低
			if buyPrice < latest.Low {
				logrus.Debugf("%v[%v][CANCEL BUY]%v < %v", i, timeStr, buyPrice, latest.Low)
				doBuyOrder()
			} else {
				base += (quote / buyPrice) * _b.GetFee()
				buyNum += 1
				logrus.Debugf("%v[%v][BUY DONE]quote:%v buyPrice:%v base:%v", i, timeStr, quote, buyPrice, base)

				quote = 0.0
				buyPrice = -1.0
			}
		}
		if sellPrice > 0 {
			// 失败, 定价太高
			if sellPrice > latest.High {
				logrus.Debugf("%v[%v][CANCEL SELL]%v > %v", i, timeStr, sellPrice, latest.High)
				doSellOrder()
			} else {
				quote += (base * sellPrice) * _b.GetFee()
				sellNum += 1
				logrus.Debugf("%v[%v][SELL DONE]base:%v sellPrice:%v quote:%v", i, timeStr, base, sellPrice, quote)

				base = 0.0
				sellPrice = -1.0
			}
		}

		if outMACDHist[i-1] < 0 && outMACDHist[i] > 0 {
			logrus.Debugf("%v[%v][GOLD]MACD:%v hist:%v", i, timeStr, outMACD[i], outMACDHist[i])
			macd = append(macd, i)
		} else if outMACDHist[i-1] > 0 && outMACDHist[i] < 0 {
			logrus.Debugf("%v[%v][DEAD]MACD:%v hist:%v", i, timeStr, outMACD[i], outMACDHist[i])
			macd = append(macd, i)
		}

		macdSize := len(macd)

		if outMACDHist[i] > 0 && quote > 0 && buyPrice < 0 {
			if macdSize > 2 && outMACD[i] > outMACD[macd[macdSize-3]] {
				lastGold := outMACD[macd[macdSize-1]]
				lastDead := outMACD[macd[macdSize-2]]
				last2Gold := outMACD[macd[macdSize-3]]
				if lastGold < 0 && lastDead > 0 && lastGold > last2Gold && outMACD[i] < 0 { // 下穿零线
					logrus.Debugf("%v[%v][SKIP BUY]quote:%v  base:%v", i, timeStr, quote, base)
				} else {
					doBuyOrder()
				}
			} else if records[i].Close > records[i].Open && records[i-1].Close > records[i-1].Open && records[i-2].Close > records[i-2].Open {
				doBuyOrder()
			}
		} else if outMACDHist[i] < 0 && base > 0 && sellPrice < 0 {

			if macdSize > 2 && outMACD[i] < outMACD[macd[macdSize-3]] {
				lastDead := outMACD[macd[macdSize-1]]
				lastGold := outMACD[macd[macdSize-2]]
				last2Dead := outMACD[macd[macdSize-3]]
				if lastDead > 0 && lastGold < 0 && lastDead < last2Dead && outMACD[i] > 0 { // 上穿零线
					logrus.Debugf("%v[%v][SKIP SELL]base:%v  quote:%v", i, timeStr, base, quote)
				} else {
					doSellOrder()
				}
			} else if records[i].Close < records[i].Open && records[i-1].Close < records[i-1].Open && records[i-2].Close < records[i-2].Open {
				doSellOrder()
			}
		}
	}
	endBase := 0.0
	endAsset := 0.0

	if _bridge != nil {
		endBase = base + quote/_records[len(_records)-1].Close
		endAsset = (quote + base*_records[len(_records)-1].Close) * _bridge[len(_bridge)-1].Close
	} else {
		endBase = base + quote/records[len(records)-1].Close
		endAsset = quote + base*records[len(records)-1].Close
	}

	startPrice := records[32].Close
	endPrice := records[len(records)-1].Close

	score := (endAsset / startAsset) / (endPrice / startPrice)
	logrus.Infof("[%v][END]score:%v startAsset:%v endAsset:%v startBase:%v endBase:%v startPrice:%v endPrice:%v sellNum:%v buyNum:%v",
		records[len(records)-1].Time.Format("01-02 15"), score, startAsset, endAsset, startBase, endBase, startPrice, endPrice, sellNum, buyNum)

	return score
}
