package exchange

import (
	"github.com/forchain/cryptotrader/binance"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/markcheno/go-talib"
	"github.com/forchain/cryptotrader/model"
	"fmt"
)

const SERVER_TIME_SPEED=time.Second

type BinanceDebugEx struct {
	Exchange
	client *binance.Client
	startTime time.Time
	lastPingTime time.Time
	lastGetRTime time.Time
	curRecords ExchangeRecords
	curMACDs MACDDatas
	serverErrorTime time.Duration
	account Account
	opts []OptRecord

	//
	serverRecords ExchangeRecords
	serverTime time.Time
	curElapsed time.Time

	initAssets model.Balance
	initBase model.Balance


	curOptIndex int64
}
func (p *BinanceDebugEx) Init(){
	p.startTime=time.Now()
	p.curOptIndex=0
	p.curElapsed=time.Now()
	log.Debugln("欢迎进入调试模式：\n登录币安官网")
	//
	log.Infof("操作频率：%s",GetOptFreTimeStr(p.CurTP.OptFrequency))
	log.Infof("止损：%f%%  止盈：%f%%",p.StopLoss*100,p.StopGain*100)
	//账户信息
	p.initAssets=model.Balance{p.CurTP.GetQuote(),10000,0}
	p.initBase=model.Balance{p.CurTP.GetBase(),0,0}
	ba:=p.initAssets
	p.account.balances=append(p.account.balances,ba)
	p.account.balances=append(p.account.balances,p.initBase)
	log.Infof("账户情况: %+v",p.account)


	//
	//加载本地的数据

	LoadData(getDataFilePath(p),&p.serverRecords,p.CurTP.Name,true)
	//p.serverRecords.Records=helpers.ReverseExchangeData(p.serverRecords.Records)
	//模拟一个服务器时间
	p.serverTime=time.Now()
	rLen:=len(p.serverRecords.Records)
	if rLen>0 {
		p.serverTime=p.serverRecords.Records[0].Time
		p.serverTime=p.serverTime.Add(getOptFreTimeDur(p.CurTP.OptFrequency)/10)
	} else {
		return
	}



	//
	log.Infof("\n当前调试服务器时间：%s",p.serverTime.Format("2006-01-02 15:04:05"))
	//
	PrintLocalRecordsInfo(p.serverRecords.Records,SHOW_MAX_NUM)

	//

	var ie IExchange=p
	tactData:=TacticsData{&p.curRecords,&p.curMACDs,ie,&p.account}
	p.Tactics.Init(&tactData)



}
func (p *BinanceDebugEx) Exit(){


	log.Infof("退出Binance市场调试模式: %s ,下次再会    o(*￣︶￣*)o",p.CurTP.Name)
}
func (p *BinanceDebugEx) GetExchange()(*Exchange){
	return &p.Exchange
}
func (p *BinanceDebugEx) Update() error  {


	if time.Since(p.curElapsed)<time.Millisecond/10 {
		return nil
	}
	rLen:=len(p.serverRecords.Records)
	if rLen<=0 {
		return fmt.Errorf("找不到调试数据，强制退出")
	}

	//
	p.curOptIndex++
	p.curElapsed=time.Now()
	p.serverTime=p.serverTime.Add(getOptFreTimeDur(p.CurTP.OptFrequency))

	log.Infof("\n\n第 %d 次操作-当前调试服务器时间：%s",p.curOptIndex,p.serverTime.Format("2006-01-02 15:04:05"))

	var err error
	//ping
	elapsed:=time.Since(p.lastPingTime)
	if elapsed>=PING_SPACE{
		err=p.ping()
		if err!=nil{
			return err
		}
	}
	//record
	err=p.updateRecords()
	if err!=nil {

		return err
	}

	//
	//调试模式下 才有  说明数据已经跑完了
	curOptTime:=getOptPreTime(p.CurTP.OptFrequency,p.serverTime)


	var lastLocalRTime time.Time

	if rLen>0 {

		lastLocalRTime = p.serverRecords.Records[rLen-1].Time
		if lastLocalRTime.Equal(curOptTime) || lastLocalRTime.Before(curOptTime) {
			p.onEndDebug()
			return fmt.Errorf("调试完毕")
		}
	}
	//
	return nil
}
func (p *BinanceDebugEx) ping() error {
	p.lastPingTime=time.Now()

	log.Infoln("Binance Ping success")

	return nil
}

func (p *BinanceDebugEx) updateRecords() error{

	var err error
	elapsed:=p.serverTime.Sub(p.lastGetRTime)
	if elapsed>=PING_SPACE{
		//看是否需要请求数据
		p.getServerTime()
		p.lastGetRTime=p.serverTime
		//
		curOptTime:=getOptPreTime(p.CurTP.OptFrequency,p.serverTime)

		var lastLocalRTime time.Time

		rLen:=len(p.curRecords.Records)
		if rLen>0 {


			lastLocalRTime=p.curRecords.Records[rLen-1].Time
			if lastLocalRTime.Equal(curOptTime)||lastLocalRTime.After(curOptTime) {
				log.Println("本地记录已经是最新")
				return nil
			}
			lastLocalRTime=lastLocalRTime.Add(-getOptFreTimeDur(p.CurTP.OptFrequency))

		} else{
			lastLocalRTime=p.serverTime
		}
		//
		err=p.getRecords(p.serverTime,lastLocalRTime)
		if err!=nil{
			return err
		}
	}

	return nil
}
func (p *BinanceDebugEx) getRecords(startTime,endTime time.Time) error {


	//
	if startTime.Equal(endTime) {
		log.Infof("向服务器请求[最开始->%s]的记录", startTime.Format("2006-01-02 15:04:05"))
	}else {
		log.Infof("向服务器请求[%s->%s]的记录",endTime.Format("2006-01-02 15:04:05"),
			startTime.Format("2006-01-02 15:04:05"))
	}
	curOptTime:=getOptPreTime(p.CurTP.OptFrequency,p.serverTime)

	curRecords:=p.serverRecords.getRecordBeforeTime(curOptTime)
	if len(curRecords)<=0{
		log.Errorf("获取记录失败：%s",curOptTime)
		return nil
	}
	//
	rLen:=len(p.curRecords.Records)
	log.Infof("新获得 %d 条数据",len(curRecords)-rLen)
	p.curRecords.Records=curRecords


	//
	p.updateMyOrders()
	//
	p.refreshMACD()
	return nil
}
func (p *BinanceDebugEx) getServerTime(){
	p.serverErrorTime=0;
	log.Infof("Binance Server time: %v  误差时间：%v秒", p.serverTime,p.serverErrorTime)
}
func (p *BinanceDebugEx)refreshMACD(){
	rLen:=len(p.curRecords.Records)
	if rLen<=30 {
		log.Errorf("     %d条记录太少了，无法产生MACD",rLen)
		return
	}
	closeRs:=p.curRecords.GetCloseRecords()


	p.curMACDs=MACDDatas{}
	outMACD, outMACDSignal, outMACDHist := talib.MacdFix(closeRs, 9)

	for  i,x:=range outMACD{
		macd:=MACDData{x,outMACDSignal[i],outMACDHist[i],i}
		p.curMACDs=append(p.curMACDs,macd)
	}

	//PrintMACDsInfo(p.curMACDs,p.curRecords.Records,SHOW_MAX_NUM*2)

	//

	//
	p.Tactics.Update()
}
func (p *BinanceDebugEx)Execute(cmd OptRecord) {

	curRecord:=p.curRecords.GetLastRecord()
	lastRecord:=p.getLastRecord()
	if curRecord==nil||lastRecord==nil {
		log.Errorf("下单失败：找不到数据")
		return
	}
	cmd.indexOpt=p.curOptIndex

	var curAPrice float64
	if cmd.price>0 {
		curAPrice=cmd.price
		log.Infof("使用自定义价格：%f   当前市价：%f",curAPrice,curRecord.Close)
	}else {
		curAPrice=curRecord.Close
	}

	switch cmd.optType {
	case binance.OrderBuy:
		log.Infof("\n建议您买入 ^_^    原因:%s",reasonStr[cmd.reason])
		//
		//定价
		//curAPrice=curRecord.Close
		/*if  curRecord.Close>curRecord.Open{
			curAPrice=getAverage(lastRecord)
		}*/
		curAPrice=getPrecisionFloat(curAPrice,p.CurTP.MinTradePrice)
		if curAPrice<=p.CurTP.MinTradePrice {
			return
		}
		balance:=p.account.getBalance(p.CurTP.GetQuote())
		if balance==nil {
			log.Errorf("下单失败：没有钱包")
			return
		}

		var canBuyNum float64=0
		free:=balance.Free

		if free>p.CurTP.MinOrderTotalPrice&&free>p.CurTP.MinTradeNum*curAPrice*(1+p.Fee) {
			canBuyNum=free/(curAPrice*(1+p.Fee))
			canBuyNum=getPrecisionFloat(canBuyNum,p.CurTP.MinTradeNum)
		}
		if canBuyNum<=p.CurTP.MinTradeNum {
			log.Infof("下单失败：您的 %s 不足，下单买入:%f",p.CurTP.GetQuote(),free)
		}else {
			amount:=canBuyNum
			cost:=amount*curAPrice+amount*curAPrice*p.Fee

			//

			free-=cost
			if free<0 {
				return
			}

			//
			balance.Free=free
			balance.Frozen+=cost
			//
			cmd.time=p.serverTime


			//order
			order:=model.Order{p.curOptIndex,amount,0,curAPrice,OrderStatus.String(OrderNew),
			binance.OrderType.String(binance.OrderLimit),binance.OrderSide.String(binance.OrderBuy),cmd.time,""}

			cmd.orderID=order.ID
			//
			p.opts=append(p.opts,cmd)
			p.account.orders=append(p.account.orders,order)

			log.Infof("成功下单：买%+v",order)
		}
	case binance.OrderSell:
		log.Infof("\n建议您买出 ^_^   原因:%s",reasonStr[cmd.reason])
		//定价
		//curAPrice=curRecord.Close
		curAPrice=getPrecisionFloat(curAPrice,p.CurTP.MinTradePrice)
		if curAPrice<=p.CurTP.MinTradePrice {
			return
		}
		balance:=p.account.getBalance(p.CurTP.GetBase())
		ownBase:=balance.Free//p.account.getOwnBase()
		openOrders:=p.account.getOpenOrders()

		if len(openOrders)>0 {
			for _,y:=range openOrders{
				x:=&p.account.orders[y]
				p.cancelOrder(x,false)
			}
		}



		if ownBase>p.CurTP.MinTradeNum*(1+p.Fee) {
			costNum:=ownBase/(1+p.Fee)
			realCostNum:=getPrecisionFloat(costNum,p.CurTP.MinTradeNum)
			sellMoney:=realCostNum*curAPrice
			if realCostNum>p.CurTP.MinTradeNum&&sellMoney>p.CurTP.MinOrderTotalPrice {
				cmd.time=p.serverTime
				//order
				order:=model.Order{p.curOptIndex,realCostNum,0,curAPrice,OrderStatus.String(OrderNew),
					binance.OrderType.String(binance.OrderLimit),binance.OrderSide.String(binance.OrderSell),cmd.time,""}

				cmd.orderID=order.ID
				//
				p.opts=append(p.opts,cmd)
				p.account.orders=append(p.account.orders,order)
				//
				log.Infof("成功下单：卖%+v",order)



			} else {
				log.Errorln("下单失败：仓位太低")
			}
		}



	default:
		log.Errorln("错误")
	}

}
func (p *BinanceDebugEx)getLastRecord() *model.Record{
	curOptT:=p.serverTime.Add(-getOptFreTimeDur(p.CurTP.OptFrequency))
	curOptTime:=getOptPreTime(p.CurTP.OptFrequency,curOptT)
	//
	return p.curRecords.getRecordByTime(curOptTime)
}
func (p *BinanceDebugEx)onEndDebug(){

	PrintMACDsInfo(p.curMACDs,p.curRecords.Records,SHOW_MAX_NUM*2)

	log.Println("所有的记录已经测试完")
	//SaveData(getDataFilePath(p),&p.curRecords,p.CurTP.Name)
	//
	PrintDebugInfo(p)


}
func (p *BinanceDebugEx)updateMyOrders() {
	//调试处理订单
	balance:=p.account.getBalance(p.CurTP.GetQuote())
	bbalance:=p.account.getBalance(p.CurTP.GetBase())

	lastRecord:=p.getLastRecord()
	if balance==nil||lastRecord==nil {
		log.Errorf("订单处理失败")
		return
	}

	openOrders:=p.account.getOpenOrders()
	if len(openOrders)>0{
		for _,y:=range openOrders{
			x:=&p.account.orders[y]
			opt:=p.getOptRecordByOrderID(x.ID)
			if opt==nil {
				continue
			}
			hourTime:=getHourTime(opt.time)
			if hourTime.Equal(lastRecord.Time)||hourTime.Before(lastRecord.Time){
				if x.Price>=lastRecord.Low&&x.Price<=lastRecord.High {

					x.DealAmount=x.Amount
					x.Status=OrderStatus.String(OrderFilled)

					if x.Side==binance.OrderSide.String(binance.OrderBuy) {
						cost:=x.Amount*x.Price*(1+p.Fee)
						//
						balance.Frozen-=cost
						if balance.Frozen<0 {
							balance.Frozen=0
						}
						bbalance.Free+=x.Amount

						//成功促成交易
						trade:=model.Trade{x.ID,TradeType.String(TradeBuy),x.Price,
						x.DealAmount,opt.time,x.Amount*x.Price*p.Fee,p.CurTP.GetQuote(),
						x.ID,""}
						p.account.trades=append(p.account.trades, trade)
						//
						log.Infof("买入 [%f] %s,花费 [%f] %s",x.DealAmount,p.CurTP.GetBase(),
							cost,p.CurTP.GetQuote())
						//


					}else {
						income:=x.Amount*x.Price
						balance.Free+=income
						bbalance.Free-=(x.Amount+x.Amount*p.Fee)

						trade:=model.Trade{x.ID,TradeType.String(TradeSell),x.Price,
							x.DealAmount,opt.time,x.Amount*x.Price*p.Fee,
							p.CurTP.GetQuote(),x.ID,""}
						p.account.trades=append(p.account.trades, trade)

						log.Infof("卖出 [%f] %s,收入 [%f] %s",x.DealAmount,p.CurTP.GetBase(),
							income,p.CurTP.GetQuote())
					}
				}else {
					p.cancelOrder(x,true)
				}
			}
		}

		//

	}


}
func (p *BinanceDebugEx)getOptRecordByOrderID(ID int64) *OptRecord {
	for _,x:=range p.opts{
		if x.orderID==ID{
			return &x
		}
	}
	return nil
}
func (p *BinanceDebugEx)cancelOrder(order *model.Order,isFail bool)  {
	if order.Status!=OrderStatus.String(OrderNew) {
		return
	}
	balance:=p.account.getBalance(p.CurTP.GetQuote())
	if isFail {
		order.Status=OrderStatus.String(OrderRejected)
	}else {
		order.Status=OrderStatus.String(OrderCanceled)
	}


	cost:=order.Amount*order.Price*(1+p.Fee)

	//
	if order.Side==binance.OrderSide.String(binance.OrderBuy) {
			//
			balance.Frozen-=cost
			balance.Free+=cost
			//

			//
	}
	log.Infof("取消订单成功:%+v",order)
}