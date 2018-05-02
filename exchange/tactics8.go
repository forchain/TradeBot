package exchange

import (
	"fmt"
	"github.com/forchain/cryptotrader/binance"
	"github.com/forchain/cryptotrader/model"
	log "github.com/sirupsen/logrus"
)

const CAP_FEE float64 = 2

func init() {
	tact := &Tactics8{}
	TacticsMap[tact.GetID()] = tact
}

type Tactics8 struct {
	tactData *TacticsData
}

func (p *Tactics8) Init(date *TacticsData) {
	p.tactData = date
	//
	log.Infof("\n*********您正使用 第%d号 策略******", p.GetID())
	//p.Update()

}
func (p *Tactics8) GetID() int {
	return 8
}
func (p *Tactics8) Update() {
	if p.tactData == nil {
		return
	}
	p.denoiseMACD()
	//

	if p.crossOverDEAByDIF() {
		return
	}
	if p.stopLoss() {
		return
	}
}
func (p *Tactics8) crossOverDEAByDIF() bool {
	mLen := len(*p.tactData.CurMACDs)
	if mLen < 6 {
		//log.Infoln("没有可用的MACD数据")
		return false
	}

	//
	tLen := len(p.tactData.account.trades)
	if tLen > 0 {
		return false
	}

	//
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
func (p *Tactics8) getLastTrendStartIndex() int {
	result := -1

	srcMACDs := *p.tactData.CurMACDs
	mLen := len(srcMACDs)
	if mLen < 6 {
		return result
	}

	meetUnTrend := 0
	//meetUnTrend:=false
	for i := mLen - 3; i >= 0; i-- {

		if meetUnTrend == 0 {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD < 0 {
				meetUnTrend = 1
			}
		} else if meetUnTrend == 1 {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD > 0 {
				meetUnTrend = 2
			}
		} else if meetUnTrend == 2 {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD < 0 {
				result = i
				break
			}
		}
		/*if meetUnTrend {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD>0 {
				result=i
				break
			}
		}else {
			if srcMACDs[mLen-2].MACD*srcMACDs[i].MACD<0 {
				meetUnTrend=true
			}
		}*/
	}
	return result
}
func (p *Tactics8) getLastHightAndLowIndex(lastIndex int) (float64, float64, error) {
	rLen := len(p.tactData.CurRecords.Records)
	if rLen <= 0 || lastIndex < 0 || lastIndex >= rLen {
		return 0, 0, fmt.Errorf("error:len")
	}
	//highAverage:=getGeometryAverage(&p.tactData.CurRecords.Records[rLen-2])
	//lowAverage:=getGeometryAverage(&p.tactData.CurRecords.Records[rLen-2])

	highAverage := p.tactData.CurRecords.Records[rLen-2].High
	lowAverage := p.tactData.CurRecords.Records[rLen-2].Low

	/*curAv:=0.0
	for i:=rLen-2;i>lastIndex ;i--  {
		curAv=getGeometryAverage(&p.tactData.CurRecords.Records[i])
		if curAv>highAverage{
			highAverage=curAv
		}
		if curAv<lowAverage {
			lowAverage=curAv
		}
	}*/

	for i := rLen - 2; i > lastIndex; i-- {
		if p.tactData.CurRecords.Records[i].High > highAverage {
			highAverage = p.tactData.CurRecords.Records[i].High
		}
		if p.tactData.CurRecords.Records[i].Low < lowAverage {
			lowAverage = p.tactData.CurRecords.Records[i].Low
		}
	}

	return highAverage, lowAverage, nil
}

func (p *Tactics8) stopLoss() bool {
	if p.processBuy() {
		return true
	}
	if p.processSell() {
		return true
	}

	return false
}
func (p *Tactics8) processBuy() bool {

	rLen := len(p.tactData.CurRecords.Records)
	rLen--
	if rLen <= 0 {
		return false
	}
	srcMACDs := *p.tactData.CurMACDs
	mLen := len(srcMACDs)
	if mLen < 6 {
		return false
	}
	////
	balance := p.tactData.account.getBalance(p.tactData.Excha.GetExchange().CurTP.GetQuote())
	if balance == nil {
		return false
	}
	free := balance.Free

	if free <= p.tactData.Excha.GetExchange().CurTP.MinOrderTotalPrice {
		return false
	}

	var trade *model.Trade
	tLen := len(p.tactData.account.trades)
	if tLen > 0 {
		/*for i:=tLen-1;i>=0 ;i--  {
			if p.tactData.account.trades[i].Type==TradeType.String(TradeSell)  {
				trade=&p.tactData.account.trades[i]
				break
			}
		}*/
		if p.tactData.account.trades[tLen-1].Type == TradeType.String(TradeSell) {
			trade = &p.tactData.account.trades[tLen-1]
		}
	}
	var tradeOptIndex int = -1
	if trade != nil {

		optTime := getOptPreTime(p.tactData.Excha.GetExchange().CurTP.OptFrequency, trade.Time)

		//
		tradeOptIndex = p.tactData.CurRecords.getRecordIndexByTime(optTime)
	}
	if tradeOptIndex == -1 {
		return false
	}
	//
	/*isFall:=false
	if  srcMACDs[mLen-2].MACD>srcMACDs[mLen-3].MACD&&
		srcMACDs[mLen-3].MACD>srcMACDs[mLen-4].MACD{
		isFall=true
	}
	if !isFall {
		return false
	}*/
	if srcMACDs[mLen-2].MACD < srcMACDs[mLen-3].MACD {
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
	//
	zhchTopPrice := (highAverage - trade.Price) * p.tactData.Excha.GetExchange().StopLoss
	if zhchTopPrice >= 0 {
		zhchTopPrice += trade.Price
	} else {
		zhchTopPrice = highAverage
	}
	//
	zhchLowPrice := (trade.Price - lowAverage) * (1 - p.tactData.Excha.GetExchange().StopLoss)
	if zhchLowPrice >= 0 {
		zhchLowPrice += lowAverage
	} else {
		zhchLowPrice = lowAverage
	}
	capPrice := (1 - p.tactData.Excha.GetExchange().Fee) * trade.Price / (1 + CAP_FEE*p.tactData.Excha.GetExchange().Fee)
	capPrice = getPrecisionFloat(capPrice, p.tactData.Excha.GetExchange().CurTP.MinTradePrice)
	//start
	var optCmd *OptRecord

	if p.tactData.CurRecords.Records[rLen-1].Close < capPrice &&
		p.tactData.CurRecords.Records[rLen-1].Close >= zhchLowPrice {

		optCmd = &OptRecord{optType: binance.OrderBuy, reason: STOP_GAIN}
	} else if p.tactData.CurRecords.Records[rLen-1].Close < zhchTopPrice &&
		p.tactData.CurRecords.Records[rLen-1].Close >= capPrice {

		optCmd = &OptRecord{optType: binance.OrderBuy, reason: STOP_GAIN, price: capPrice}
	} else if p.tactData.CurRecords.Records[rLen-1].Close >= zhchTopPrice {
		optCmd = &OptRecord{optType: binance.OrderBuy, reason: STOP_GAIN}
	}

	if optCmd != nil {
		p.tactData.Excha.Execute(*optCmd)
		return true
	}
	return false
}
func (p *Tactics8) processSell() bool {

	rLen := len(p.tactData.CurRecords.Records)
	rLen--
	if rLen <= 0 {
		return false
	}
	srcMACDs := *p.tactData.CurMACDs
	mLen := len(srcMACDs)
	if mLen < 6 {
		return false
	}
	////
	ownBaseFree := 0.0
	ownBase := p.tactData.account.getBalance(p.tactData.Excha.GetExchange().CurTP.GetBase())
	if ownBase != nil {
		ownBaseFree = ownBase.Free
	}
	income := ownBaseFree - ownBaseFree*p.tactData.Excha.GetExchange().Fee
	if income <= 0 || income <= p.tactData.Excha.GetExchange().CurTP.MinTradeNum {
		return false
	}

	var trade *model.Trade
	tLen := len(p.tactData.account.trades)
	if tLen > 0 {
		/*for i:=tLen-1;i>=0 ;i--  {
			if p.tactData.account.trades[i].Type==TradeType.String(TradeBuy)  {
				trade=&p.tactData.account.trades[i]
				break
			}
		}*/

		if p.tactData.account.trades[tLen-1].Type == TradeType.String(TradeBuy) {
			trade = &p.tactData.account.trades[tLen-1]
		}
	}
	var tradeOptIndex int = -1
	if trade != nil {

		optTime := getOptPreTime(p.tactData.Excha.GetExchange().CurTP.OptFrequency, trade.Time)

		//
		tradeOptIndex = p.tactData.CurRecords.getRecordIndexByTime(optTime)
	}
	if tradeOptIndex == -1 {
		return false
	}
	//
	/*isFall:=false
	if  srcMACDs[mLen-2].MACD<srcMACDs[mLen-3].MACD&&
		srcMACDs[mLen-3].MACD<srcMACDs[mLen-4].MACD{
		isFall=true
	}
	if !isFall {
		return false
	}*/
	if srcMACDs[mLen-2].MACD > srcMACDs[mLen-3].MACD {
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
	//
	zhchTopPrice := (highAverage - trade.Price) * p.tactData.Excha.GetExchange().StopLoss
	if zhchTopPrice >= 0 {
		zhchTopPrice += trade.Price
	} else {
		zhchTopPrice = highAverage
	}
	//
	zhchLowPrice := (trade.Price - lowAverage) * (1 - p.tactData.Excha.GetExchange().StopLoss)
	if zhchLowPrice >= 0 {
		zhchLowPrice += lowAverage
	} else {
		zhchLowPrice = lowAverage
	}
	capPrice := (1 + p.tactData.Excha.GetExchange().Fee) * trade.Price / (1 - CAP_FEE*p.tactData.Excha.GetExchange().Fee)
	capPrice = getPrecisionFloat(capPrice, p.tactData.Excha.GetExchange().CurTP.MinTradePrice)
	//start
	var optCmd *OptRecord

	if p.tactData.CurRecords.Records[rLen-1].Close < zhchTopPrice &&
		p.tactData.CurRecords.Records[rLen-1].Close > capPrice {

		optCmd = &OptRecord{optType: binance.OrderSell, reason: STOP_LOSS}
	} else if p.tactData.CurRecords.Records[rLen-1].Close <= capPrice &&
		p.tactData.CurRecords.Records[rLen-1].Close > zhchLowPrice {

		optCmd = &OptRecord{optType: binance.OrderSell, reason: STOP_LOSS, price: capPrice}
	} else if p.tactData.CurRecords.Records[rLen-1].Close <= zhchLowPrice {
		optCmd = &OptRecord{optType: binance.OrderSell, reason: STOP_LOSS}
	}

	if optCmd != nil {
		p.tactData.Excha.Execute(*optCmd)
		return true
	}
	return false
}

/*MACD的降噪处理*/
func (p *Tactics8) denoiseMACD() {
	//copy
	mLen := len(*p.tactData.CurMACDs)
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
