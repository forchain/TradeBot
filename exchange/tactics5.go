package exchange

import (
	"github.com/forchain/cryptotrader/binance"
	log "github.com/sirupsen/logrus"
)

func init() {
	TacticsMap[5] = &Tactics5{}
}

type Tactics5 struct {
	tactData *TacticsData
}

func (p *Tactics5) Init(date *TacticsData) {
	p.tactData = date
	//
	log.Infof("\n*********您正使用 第%d号 策略******", p.GetID())
	//p.Update()

}
func (p *Tactics5) GetID() int {
	return 5
}
func (p *Tactics5) Update() {
	if p.tactData == nil {
		return
	}

	//

	if p.crossOver0ByDIF() {
		return
	}
	if p.crossOverDEAByDIF() {
		return
	}
	if p.stopLoss() {
		return
	}
}

/*DIF突破0轴*/
func (p *Tactics5) crossOver0ByDIF() bool {
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
	/*	if srcMACDs[mLen-4].DIF<0&&srcMACDs[mLen-3].DIF<0&&srcMACDs[mLen-2].DIF>0&&srcMACDs[mLen-1].DIF>0&&
		srcMACDs[mLen-4].DIF<srcMACDs[mLen-3].DIF&&srcMACDs[mLen-3].DIF<srcMACDs[mLen-2].DIF{

		optCmd:=OptRecord{optType:binance.OrderBuy,reason:DIF_UP_0}
		p.tactData.Excha.Execute(optCmd)
		return true
	}*/

	if srcMACDs[mLen-4].DIF > 0 && srcMACDs[mLen-3].DIF > 0 && srcMACDs[mLen-2].DIF < 0 && srcMACDs[mLen-1].DIF < 0 &&
		srcMACDs[mLen-4].DIF > srcMACDs[mLen-3].DIF && srcMACDs[mLen-3].DIF > srcMACDs[mLen-2].DIF {
		optCmd := OptRecord{optType: binance.OrderSell, reason: DIF_DOWN_0}
		p.tactData.Excha.Execute(optCmd)
		return true
	}
	return false
}
func (p *Tactics5) crossOverDEAByDIF() bool {
	mLen := len(*p.tactData.CurMACDs)
	if mLen < 6 {
		//log.Infoln("没有可用的MACD数据")
		return false
	}

	srcMACDs := *p.tactData.CurMACDs

	//寻找死叉
	//3+2连续
	if mLen >= 6 {
		if srcMACDs[mLen-2].MACD < 0 &&
			srcMACDs[mLen-3].MACD < 0 &&
			srcMACDs[mLen-4].MACD > 0 &&
			srcMACDs[mLen-5].MACD > 0 &&
			srcMACDs[mLen-6].MACD > 0 {
			optCmd := OptRecord{optType: binance.OrderSell, reason: DEAD_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	//3+3连续
	if mLen >= 7 {
		if srcMACDs[mLen-2].MACD < 0 &&
			srcMACDs[mLen-3].MACD < 0 &&
			srcMACDs[mLen-4].MACD < 0 &&
			srcMACDs[mLen-5].MACD > 0 &&
			srcMACDs[mLen-6].MACD > 0 &&
			srcMACDs[mLen-7].MACD > 0 {
			optCmd := OptRecord{optType: binance.OrderSell, reason: DEAD_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	//3+4连续
	if mLen >= 8 {
		if srcMACDs[mLen-2].MACD < 0 &&
			srcMACDs[mLen-3].MACD < 0 &&
			srcMACDs[mLen-4].MACD < 0 &&
			srcMACDs[mLen-5].MACD < 0 &&
			srcMACDs[mLen-6].MACD > 0 &&
			srcMACDs[mLen-7].MACD > 0 &&
			srcMACDs[mLen-8].MACD > 0 {
			optCmd := OptRecord{optType: binance.OrderSell, reason: DEAD_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	//3+5连续
	if mLen >= 9 {
		if srcMACDs[mLen-2].MACD < 0 &&
			srcMACDs[mLen-3].MACD < 0 &&
			srcMACDs[mLen-4].MACD < 0 &&
			srcMACDs[mLen-5].MACD < 0 &&
			srcMACDs[mLen-6].MACD < 0 &&
			srcMACDs[mLen-7].MACD > 0 &&
			srcMACDs[mLen-8].MACD > 0 &&
			srcMACDs[mLen-9].MACD > 0 {
			optCmd := OptRecord{optType: binance.OrderSell, reason: DEAD_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}

	//寻找金叉
	//3+3连续
	if mLen >= 7 {
		if srcMACDs[mLen-2].MACD > 0 &&
			srcMACDs[mLen-3].MACD > 0 &&
			srcMACDs[mLen-4].MACD > 0 &&
			srcMACDs[mLen-5].MACD < 0 &&
			srcMACDs[mLen-6].MACD < 0 &&
			srcMACDs[mLen-7].MACD < 0 {
			optCmd := OptRecord{optType: binance.OrderBuy, reason: GOLDEN_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	//3+4连续
	if mLen >= 8 {
		if srcMACDs[mLen-2].MACD > 0 &&
			srcMACDs[mLen-3].MACD > 0 &&
			srcMACDs[mLen-4].MACD > 0 &&
			srcMACDs[mLen-5].MACD > 0 &&
			srcMACDs[mLen-6].MACD < 0 &&
			srcMACDs[mLen-7].MACD < 0 &&
			srcMACDs[mLen-8].MACD < 0 {
			optCmd := OptRecord{optType: binance.OrderBuy, reason: GOLDEN_CROSS}
			p.tactData.Excha.Execute(optCmd)
			return true

		}
	}
	return false
}
func (p *Tactics5) stopLoss() bool {
	ownBase := p.tactData.account.getOwnBase()
	if ownBase <= 0 {
		return false
	}
	tLen := len(p.tactData.account.trades)
	if tLen <= 0 {
		return false
	}
	if p.tactData.account.trades[tLen-1].Type != TradeType.String(TradeBuy) {
		return false
	}
	rLen := len(p.tactData.CurRecords.Records)
	if rLen <= 0 {
		return false
	}
	if p.tactData.CurRecords.Records[rLen-1].Close*1.1 < p.tactData.account.trades[tLen-1].Price {

		optCmd := OptRecord{optType: binance.OrderSell, reason: STOP_LOSS}
		p.tactData.Excha.Execute(optCmd)

		return true
	}
	return false
}
