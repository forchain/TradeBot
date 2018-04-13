package exchange

import (
	log "github.com/sirupsen/logrus"
	"github.com/forchain/cryptotrader/binance"
)

type Tactics2 struct {
	tactData *TacticsData
}
func (p *Tactics2)Init(date *TacticsData){
	p.tactData=date
	log.Infof("\n*********您正使用 第%d号 策略******",p.GetID())
	//

	//p.Update()

}
func (p *Tactics2)GetID() int{
	return 2
}
func (p *Tactics2)Update(){
	if p.tactData==nil {
		return
	}
	p.denoiseMACD()

	//
	if p.crossOverByDIF(){
		return
	}
	if p.fallAwayByDIF(){
		return
	}
}
/*MACD的降噪处理*/
func (p *Tactics2) denoiseMACD(){
	//copy
	mLen:=len(*p.tactData.CurMACDs)
	if mLen<=5 {
		log.Infoln("没有可用的MACD数据")
		return
	}
	//log.Infoln("正在进行MACD降噪处理...")
	var optMACDs MACDDatas
	//1
	vmLen:=mLen-1
	srcMACDs:=*p.tactData.CurMACDs
	for i:=0;i<vmLen;i++ {
		if i>=2&&(i+2)<vmLen {
			if srcMACDs[i-2].MACD>0&&srcMACDs[i-1].MACD>0&&
				srcMACDs[i].MACD<0&&
				srcMACDs[i+1].MACD>0&&srcMACDs[i+2].MACD>0  {

				srcMACDs[i].MACD=0
				srcMACDs[i].DIF=srcMACDs[i].DEA+0.01
				//
				optMACDs=append(optMACDs,srcMACDs[i])
			}
			if srcMACDs[i-2].MACD<0&&srcMACDs[i-1].MACD<0&&
				srcMACDs[i].MACD>0&&
				srcMACDs[i+1].MACD<0&&srcMACDs[i+2].MACD<0  {
				srcMACDs[i].MACD=0
				srcMACDs[i].DIF=srcMACDs[i].DEA-0.01
				//
				optMACDs=append(optMACDs,srcMACDs[i])
			}
		}
	}
	//2
	for i:=0;i<vmLen;i++ {
		if i>=3&&(i+4)<vmLen {
			//可以处理
			if srcMACDs[i-3].MACD>0&&srcMACDs[i-2].MACD>0&&srcMACDs[i-1].MACD>0&&
				srcMACDs[i].MACD<0&&srcMACDs[i+1].MACD<0&&
				srcMACDs[i+2].MACD>0&&srcMACDs[i+3].MACD>0&&srcMACDs[i+4].MACD>0{

				srcMACDs[i].MACD=0
				srcMACDs[i].DIF=srcMACDs[i].DEA+0.01

				srcMACDs[i+1].MACD=0
				srcMACDs[i+1].DIF=srcMACDs[i+1].DEA+0.01
				//
				optMACDs=append(optMACDs,srcMACDs[i])
				optMACDs=append(optMACDs,srcMACDs[i+1])
			}
			if srcMACDs[i-3].MACD<0&&srcMACDs[i-2].MACD<0&&srcMACDs[i-1].MACD<0&&
				srcMACDs[i].MACD>0&&srcMACDs[i+1].MACD>0&&
				srcMACDs[i+2].MACD<0&&srcMACDs[i+3].MACD<0&&srcMACDs[i+4].MACD<0{

				srcMACDs[i].MACD=0
				srcMACDs[i].DIF=srcMACDs[i].DEA-0.01

				srcMACDs[i+1].MACD=0
				srcMACDs[i+1].DIF=srcMACDs[i+1].DEA-0.01
				//
				optMACDs=append(optMACDs,srcMACDs[i])
				optMACDs=append(optMACDs,srcMACDs[i+1])
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
	if len(optMACDs)==0 {
		return
	}
	PrintOptimalMACDsInfo(optMACDs,(*p.tactData.CurRecords).Records,false)
}
/*DIF突破0轴*/
func (p *Tactics2) crossOverByDIF() bool{
	mLen:=len(*p.tactData.CurMACDs)
	if mLen<=0 {
		log.Infoln("没有可用的MACD数据")
		return false
	}

	srcMACDs:=*p.tactData.CurMACDs

	if mLen<4 {
		return false
	}
	if srcMACDs[mLen-4].DIF<0&&srcMACDs[mLen-3].DIF<0&&srcMACDs[mLen-2].DIF>0&&srcMACDs[mLen-1].DIF>0&&
		srcMACDs[mLen-4].DIF<srcMACDs[mLen-3].DIF&&srcMACDs[mLen-3].DIF<srcMACDs[mLen-2].DIF{

		optCmd:=OptRecord{optType:binance.OrderBuy,reason:DIF_UP_0}
		p.tactData.Excha.Execute(optCmd)
		return true
	}

	if srcMACDs[mLen-4].DIF>0&&srcMACDs[mLen-3].DIF>0&&srcMACDs[mLen-2].DIF<0&&srcMACDs[mLen-1].DIF<0&&
		srcMACDs[mLen-4].DIF>srcMACDs[mLen-3].DIF&&srcMACDs[mLen-3].DIF>srcMACDs[mLen-2].DIF{
		optCmd:=OptRecord{optType:binance.OrderSell,reason:DIF_DOWN_0}
		p.tactData.Excha.Execute(optCmd)
		return true
	}
	return false
}
/*DIF背离*/
func (p *Tactics2) fallAwayByDIF() bool{
	mLen:=len(*p.tactData.CurMACDs)
	if mLen<=0 {
		log.Infoln("没有可用的MACD数据")
		return false
	}

	srcMACDs:=*p.tactData.CurMACDs

	var optMACDs MACDDatas

	//寻找死叉 目前定位3个连续
	optMACDs=MACDDatas{}


	for i:=mLen-1; i>=5;i--  {
		if srcMACDs[i].DIF<srcMACDs[i].DEA&&
			srcMACDs[i-1].DIF<srcMACDs[i-1].DEA&&
			srcMACDs[i-2].DIF<srcMACDs[i-2].DEA&&
			srcMACDs[i-3].DIF>srcMACDs[i-3].DEA&&
			srcMACDs[i-4].DIF>srcMACDs[i-4].DEA&&
			srcMACDs[i-5].DIF>srcMACDs[i-5].DEA{

			if srcMACDs[mLen-1].DIF>=0 {
				if srcMACDs[i-2].DIF>0&&srcMACDs[i-3].DIF>0{
					optMACDs=append(optMACDs,srcMACDs[i-2])
					optMACDs=append(optMACDs,srcMACDs[i-3])
					break
				}
			}else if srcMACDs[mLen-1].DIF<=0{
				if srcMACDs[i-2].DIF<0&&srcMACDs[i-3].DIF<0{
					optMACDs=append(optMACDs,srcMACDs[i-2])
					optMACDs=append(optMACDs,srcMACDs[i-3])
					break
				}
			}

			break
		}
	}
	if len(optMACDs)==2 {
		if srcMACDs[mLen-1].MACD>0 {
			if srcMACDs[mLen-1].DIF>optMACDs[0].DIF {

				optCmd:=OptRecord{optType:binance.OrderBuy,reason:FALLAWAY_BOTTOM}
				p.tactData.Excha.Execute(optCmd)
				return true
			}
		}else {
			if srcMACDs[mLen-1].DIF>optMACDs[1].DIF {

				optCmd:=OptRecord{optType:binance.OrderBuy,reason:FALLAWAY_BOTTOM}
				p.tactData.Excha.Execute(optCmd)
				return true
			}
		}
	}

	//寻找金叉 目前定位3个连续
	optMACDs=MACDDatas{}

	for i:=mLen-1; i>=5;i--  {
		if srcMACDs[i].DIF>srcMACDs[i].DEA&&
			srcMACDs[i-1].DIF>srcMACDs[i-1].DEA&&
			srcMACDs[i-2].DIF>srcMACDs[i-2].DEA&&
			srcMACDs[i-3].DIF<srcMACDs[i-3].DEA&&
			srcMACDs[i-4].DIF<srcMACDs[i-4].DEA&&
			srcMACDs[i-5].DIF<srcMACDs[i-5].DEA{


			if srcMACDs[mLen-1].DIF>=0 {
				if srcMACDs[i-2].DIF>0&&srcMACDs[i-3].DIF>0{
					optMACDs=append(optMACDs,srcMACDs[i-2])
					optMACDs=append(optMACDs,srcMACDs[i-3])
					break
				}
			}else if srcMACDs[mLen-1].DIF<=0{
				if srcMACDs[i-2].DIF<0&&srcMACDs[i-3].DIF<0{
					optMACDs=append(optMACDs,srcMACDs[i-2])
					optMACDs=append(optMACDs,srcMACDs[i-3])
					break
				}
			}

			break
		}
	}
	if len(optMACDs)==2 {
		if srcMACDs[mLen-1].MACD>0 {
			if srcMACDs[mLen-1].DIF<optMACDs[1].DIF {

				optCmd:=OptRecord{optType:binance.OrderSell,reason:FALLAWAY_TOP}
				p.tactData.Excha.Execute(optCmd)
				return true
			}
		}else {
			if srcMACDs[mLen-1].DIF>optMACDs[0].DIF {

				optCmd:=OptRecord{optType:binance.OrderSell,reason:FALLAWAY_TOP}
				p.tactData.Excha.Execute(optCmd)
				return true
			}
		}
	}
	return false
}
