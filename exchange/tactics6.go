package exchange

import (
	"fmt"
	"github.com/forchain/cryptotrader/binance"
	"github.com/forchain/cryptotrader/model"
	log "github.com/sirupsen/logrus"
)

func init() {
	TacticsMap[6] = &Tactics6{}
}

type Tactics6 struct {
	tactData *TacticsData
}

func (p *Tactics6) Init(date *TacticsData) {
	p.tactData = date
	//
	log.Infof("\n*********您正使用 第%d号 策略******", p.GetID())
	//p.Update()

}
func (p *Tactics6) GetID() int {
	return 6
}
func (p *Tactics6) Update() {
	if p.tactData == nil {
		return
	}
	p.denoiseMACD()
	//
	if p.stopLoss() {
		return
	}
	if p.crossOver0ByDIF() {
		return
	}
	if p.crossOverDEAByDIF() {
		return
	}

}

/*DIF突破0轴*/
func (p *Tactics6) crossOver0ByDIF() bool {
	mLen := len(*p.tactData.CurMACDs)
	//过滤掉最后一个数据
	mLen--
	//-------------
	if mLen <= 0 {
		log.Infoln("没有可用的MACD数据")
		return false
	}

	srcMACDs := *p.tactData.CurMACDs

	if mLen < 4 {
		return false
	}
	if srcMACDs[mLen-4].DIF < 0 && srcMACDs[mLen-3].DIF < 0 && srcMACDs[mLen-2].DIF > 0 && srcMACDs[mLen-1].DIF > 0 &&
		srcMACDs[mLen-4].DIF < srcMACDs[mLen-3].DIF && srcMACDs[mLen-3].DIF < srcMACDs[mLen-2].DIF {

		optCmd := OptRecord{optType: binance.OrderBuy, reason: DIF_UP_0}
		p.tactData.Excha.Execute(optCmd)
		return true
	}

	/*if srcMACDs[mLen-4].DIF>0&&srcMACDs[mLen-3].DIF>0&&srcMACDs[mLen-2].DIF<0&&srcMACDs[mLen-1].DIF<0&&
		srcMACDs[mLen-4].DIF>srcMACDs[mLen-3].DIF&&srcMACDs[mLen-3].DIF>srcMACDs[mLen-2].DIF{
		optCmd:=OptRecord{optType:binance.OrderSell,reason:DIF_DOWN_0}
		p.tactData.Excha.Execute(optCmd)
		return true
	}*/
	return false
}
func (p *Tactics6) crossOverDEAByDIF() bool {
	mLen := len(*p.tactData.CurMACDs)
	if mLen < 6 {
		//log.Infoln("没有可用的MACD数据")
		return false
	}

	srcMACDs := *p.tactData.CurMACDs

	//寻找金叉
	//3+3连续
	if mLen >= 7 {
		if srcMACDs[mLen-2].MACD > srcMACDs[mLen-3].MACD &&
			srcMACDs[mLen-3].MACD > srcMACDs[mLen-4].MACD &&
			srcMACDs[mLen-4].MACD > 0 &&
			srcMACDs[mLen-5].MACD < 0 &&
			srcMACDs[mLen-6].MACD < 0 &&
			srcMACDs[mLen-7].MACD < 0 {
			optCmd := OptRecord{optType: binance.OrderBuy, reason: GOLDEN_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	//4+2连续
	if mLen >= 7 {
		if srcMACDs[mLen-2].MACD > srcMACDs[mLen-3].MACD &&
			srcMACDs[mLen-3].MACD > 0 &&
			srcMACDs[mLen-4].MACD < 0 &&
			srcMACDs[mLen-5].MACD < srcMACDs[mLen-4].MACD &&
			srcMACDs[mLen-6].MACD < 0 &&
			srcMACDs[mLen-7].MACD < 0 {
			optCmd := OptRecord{optType: binance.OrderBuy, reason: GOLDEN_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	return false
}
func (p *Tactics6) getLastTrendStartIndex() int {
	result := -1

	srcMACDs := *p.tactData.CurMACDs
	mLen := len(srcMACDs)
	if mLen < 6 {
		return result
	}

	meetUnTrend := false
	for i := mLen - 3; i >= 0; i-- {

		if meetUnTrend {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD > 0 {
				result = i
				break
			}
		} else {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD < 0 {
				meetUnTrend = true
			}
		}
	}
	return result
}
func (p *Tactics6) getLastHightAndLowIndex(lastIndex int) (float64, float64, error) {
	rLen := len(p.tactData.CurRecords.Records)
	if rLen <= 0 || lastIndex < 0 || lastIndex >= rLen {
		return 0, 0, fmt.Errorf("error:len")
	}
	highAverage := getGeometryAverage(&p.tactData.CurRecords.Records[rLen-2])
	lowAverage := getGeometryAverage(&p.tactData.CurRecords.Records[rLen-2])
	curAv := 0.0
	for i := rLen - 2; i > lastIndex; i-- {
		curAv = getGeometryAverage(&p.tactData.CurRecords.Records[i])
		if curAv > highAverage {
			highAverage = curAv
		}
		if curAv < lowAverage {
			lowAverage = curAv
		}
	}
	return highAverage, lowAverage, nil
}

func (p *Tactics6) stopLoss() bool {
	//
	ownBaseFree := 0.0
	ownBase := p.tactData.account.getBalance(p.tactData.Excha.GetExchange().CurTP.GetBase())
	if ownBase != nil {
		ownBaseFree = ownBase.Free
	}
	income := ownBaseFree - ownBaseFree*p.tactData.Excha.GetExchange().Fee
	if income <= 0 || income <= p.tactData.Excha.GetExchange().CurTP.MinTradeNum {
		return false
	}
	rLen := len(p.tactData.CurRecords.Records)
	if rLen <= 0 {
		return false
	}
	srcMACDs := *p.tactData.CurMACDs
	mLen := len(srcMACDs)
	if mLen < 6 {
		return false
	}

	var trade *model.Trade
	tLen := len(p.tactData.account.trades)
	if tLen > 0 {
		for i := tLen - 1; i >= 0; i-- {
			if p.tactData.account.trades[i].Type == TradeType.String(TradeBuy) {
				trade = &p.tactData.account.trades[i]
				break
			}
		}
	}
	var tradeOptIndex int = -1
	if trade != nil {

		optTime := getOptPreTime(p.tactData.Excha.GetExchange().CurTP.OptFrequency, trade.Time)

		//
		tradeOptIndex = p.tactData.CurRecords.getRecordIndexByTime(optTime)
	}
	//
	isFall := false
	if srcMACDs[mLen-2].MACD < srcMACDs[mLen-3].MACD &&
		srcMACDs[mLen-3].MACD < srcMACDs[mLen-4].MACD {
		isFall = true
	}
	if !isFall {
		return false
	}
	lastTSIndex := p.getLastTrendStartIndex()
	if lastTSIndex < 0 {
		return false
	}
	highAverage, lowAverage, err := p.getLastHightAndLowIndex(lastTSIndex)
	if err != nil {
		return false
	}
	if highAverage < lowAverage {
		highAverage, lowAverage = lowAverage, highAverage
	}
	optPrice := 0.0
	if tradeOptIndex == -1 {
		optPrice = (highAverage-lowAverage)*(1-p.tactData.Excha.GetExchange().StopLoss) + lowAverage
	} else {

		//tradeRecord:=p.tactData.CurRecords.Records[tradeOptIndex]
		tradeMACD := srcMACDs[tradeOptIndex]
		if tradeOptIndex > lastTSIndex {
			if tradeMACD.MACD*srcMACDs[mLen-2].MACD >= 0 {
				optPrice = (highAverage-lowAverage)*(1-p.tactData.Excha.GetExchange().StopGain) + lowAverage
			} else {
				if trade.Price > lowAverage {
					lowAverage = trade.Price
				}
				if trade.Price > highAverage {
					highAverage = trade.Price
				}
				optPrice = (highAverage-lowAverage)*(1-p.tactData.Excha.GetExchange().StopGain) + lowAverage
			}
		} else {
			if trade.Price > lowAverage {
				lowAverage = trade.Price
			}
			if trade.Price > highAverage {
				highAverage = trade.Price
			}
			optPrice = (highAverage-lowAverage)*(1-p.tactData.Excha.GetExchange().StopGain) + lowAverage
		}

	}
	if p.tactData.CurRecords.Records[rLen-1].Open <= optPrice {
		optCmd := OptRecord{optType: binance.OrderSell, reason: STOP_LOSS}
		p.tactData.Excha.Execute(optCmd)
		return true
	}

	return false
}

/*MACD的降噪处理*/
func (p *Tactics6) denoiseMACD() {
	//copy
	mLen := len(*p.tactData.CurMACDs)
	//-------------
	if mLen <= 5 {
		log.Infoln("没有可用的MACD数据")
		return
	}
	//log.Infoln("正在进行MACD降噪处理...")
	var optMACDs MACDDatas
	//1
	vmLen := mLen - 1
	srcMACDs := *p.tactData.CurMACDs
	for i := 0; i < vmLen; i++ {
		if i >= 2 && (i+2) < vmLen {
			if srcMACDs[i-2].MACD > 0 && srcMACDs[i-1].MACD > 0 &&
				srcMACDs[i].MACD < 0 &&
				srcMACDs[i+1].MACD > 0 && srcMACDs[i+2].MACD > 0 {

				srcMACDs[i].MACD = 0
				srcMACDs[i].DIF = srcMACDs[i].DEA + 0.01
				//
				optMACDs = append(optMACDs, srcMACDs[i])
			}
			if srcMACDs[i-2].MACD < 0 && srcMACDs[i-1].MACD < 0 &&
				srcMACDs[i].MACD > 0 &&
				srcMACDs[i+1].MACD < 0 && srcMACDs[i+2].MACD < 0 {
				srcMACDs[i].MACD = 0
				srcMACDs[i].DIF = srcMACDs[i].DEA - 0.01
				//
				optMACDs = append(optMACDs, srcMACDs[i])
			}
		}
	}
	//2
	for i := 0; i < vmLen; i++ {
		if i >= 3 && (i+4) < vmLen {
			//可以处理
			if srcMACDs[i-3].MACD > 0 && srcMACDs[i-2].MACD > 0 && srcMACDs[i-1].MACD > 0 &&
				srcMACDs[i].MACD < 0 && srcMACDs[i+1].MACD < 0 &&
				srcMACDs[i+2].MACD > 0 && srcMACDs[i+3].MACD > 0 && srcMACDs[i+4].MACD > 0 {

				srcMACDs[i].MACD = 0
				srcMACDs[i].DIF = srcMACDs[i].DEA + 0.01

				srcMACDs[i+1].MACD = 0
				srcMACDs[i+1].DIF = srcMACDs[i+1].DEA + 0.01
				//
				optMACDs = append(optMACDs, srcMACDs[i])
				optMACDs = append(optMACDs, srcMACDs[i+1])
			}
			if srcMACDs[i-3].MACD < 0 && srcMACDs[i-2].MACD < 0 && srcMACDs[i-1].MACD < 0 &&
				srcMACDs[i].MACD > 0 && srcMACDs[i+1].MACD > 0 &&
				srcMACDs[i+2].MACD < 0 && srcMACDs[i+3].MACD < 0 && srcMACDs[i+4].MACD < 0 {

				srcMACDs[i].MACD = 0
				srcMACDs[i].DIF = srcMACDs[i].DEA - 0.01

				srcMACDs[i+1].MACD = 0
				srcMACDs[i+1].DIF = srcMACDs[i+1].DEA - 0.01
				//
				optMACDs = append(optMACDs, srcMACDs[i])
				optMACDs = append(optMACDs, srcMACDs[i+1])
			}
		}
	}
	//3
	/*for i:=0;i<vmLen;i++ {
		if i>=4&&(i+6)<vmLen {
			//可以处理
			if srcMACDs[i-4].MACD>0&&srcMACDs[i-3].MACD>0&&srcMACDs[i-2].MACD>0&&srcMACDs[i-1].MACD>0&&
				srcMACDs[i].MACD<0&&srcMACDs[i+1].MACD<0&&srcMACDs[i+2].MACD<0&&
				srcMACDs[i+3].MACD>0&&srcMACDs[i+4].MACD>0&&srcMACDs[i+5].MACD>0&&srcMACDs[i+6].MACD>0{

				srcMACDs[i].MACD=0
				srcMACDs[i].DIF=srcMACDs[i].DEA+0.01

				srcMACDs[i+1].MACD=0
				srcMACDs[i+1].DIF=srcMACDs[i+1].DEA+0.01

				srcMACDs[i+2].MACD=0
				srcMACDs[i+2].DIF=srcMACDs[i+2].DEA+0.01
				//
				optMACDs=append(optMACDs,srcMACDs[i])
				optMACDs=append(optMACDs,srcMACDs[i+1])
				optMACDs=append(optMACDs,srcMACDs[i+2])
			}
			if srcMACDs[i-4].MACD<0&&srcMACDs[i-3].MACD<0&&srcMACDs[i-2].MACD<0&&srcMACDs[i-1].MACD<0&&
				srcMACDs[i].MACD>0&&srcMACDs[i+1].MACD>0&&srcMACDs[i+2].MACD>0&&
				srcMACDs[i+3].MACD<0&&srcMACDs[i+4].MACD<0&&srcMACDs[i+5].MACD<0&&srcMACDs[i+6].MACD<0{

				srcMACDs[i].MACD=0
				srcMACDs[i].DIF=srcMACDs[i].DEA-0.01

				srcMACDs[i+1].MACD=0
				srcMACDs[i+1].DIF=srcMACDs[i+1].DEA-0.01

				srcMACDs[i+2].MACD=0
				srcMACDs[i+2].DIF=srcMACDs[i+2].DEA-0.01
				//
				optMACDs=append(optMACDs,srcMACDs[i])
				optMACDs=append(optMACDs,srcMACDs[i+1])
				optMACDs=append(optMACDs,srcMACDs[i+2])
			}
		}
	}*/
	//
	if len(optMACDs) == 0 {
		return
	}
	PrintOptimalMACDsInfo(optMACDs, (*p.tactData.CurRecords).Records, false)
}
