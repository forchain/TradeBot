package exchange

import (
	"github.com/forchain/cryptotrader/binance"
	"context"
	"time"
	log "github.com/sirupsen/logrus"
	"fmt"
	"github.com/markcheno/go-talib"
	"github.com/forchain/cryptotrader/model"
)

var PING_SPACE time.Duration
const TIMEOUT_TIME=10*time.Second

const SHOW_MAX_NUM=20

type BinanceEx struct {
	Exchange
	client *binance.Client

	lastPingTime time.Time
	lastGetRTime time.Time
	curRecords ExchangeRecords
	curMACDs MACDDatas
	serverErrorTime time.Duration
	account Account

	curOptIndex int64

	opts []OptRecord

	initAssets model.Balance
	initBase model.Balance

	initExRate float64
}
func (p *BinanceEx) Init(){
	p.StartTime=time.Now()
	p.curOptIndex=0

	log.Infof("操作频率：%s",GetOptFreTimeStr(p.CurTP.OptFrequency))
	PING_SPACE=getOptFreTimeDur(p.CurTP.OptFrequency)

	//BINANCE_APIKEY:= os.Getenv("BINANCE_APIKEY")
	//BINANCE_SECRETKEY:=os.Getenv("BINANCE_SECRETKEY")
	if len(p.APIKey)<=0||len(p.APISecretKey)<=0 {
		log.Errorf("API无账户配置")
		return
	}
	p.client= binance.New(p.APIKey, p.APISecretKey)
	log.Infof("key:%s  secret:%s",p.APIKey,p.APISecretKey)


	//
	//加载本地的数据

	LoadData(getDataFileName(p),&p.curRecords)

	PrintLocalRecordsInfo(p.curRecords.Records,SHOW_MAX_NUM)

	var ie IExchange=p
	tactData:=TacticsData{&p.curRecords,&p.curMACDs,ie,&p.account}
	p.Tactics.Init(&tactData)
}
func (p *BinanceEx) Exit(){


	log.Infoln("退出Binance市场")
}
func (p *BinanceEx) GetExchange()(*Exchange){
	return &p.Exchange
}
func (p *BinanceEx) Update() error  {

	time.Sleep(time.Second)
	var err error
	if p.client==nil {
		return fmt.Errorf("错误：client为空")
	}
	//ping
	elapsed:=time.Since(p.lastPingTime)
	if elapsed>=PING_SPACE{
		err=p.ping()
		if err!=nil{
			return err
		}
	}
	//

	elapsed=time.Since(p.lastGetRTime)
	if elapsed>=PING_SPACE{

		p.curOptIndex++
		log.Infof("\n\n第 %d 次操作",p.curOptIndex)

		p.lastGetRTime=time.Now()
		//
		err=p.getServerTime()
		if err!=nil {
			return err
		}
		//
		p.updateOrders()
		p.getMyTrades()
		//
		err=p.getAccount()
		if err!=nil {
			return err
		}
		err=p.getTicker()
		if err!=nil {
			return err
		}
		//record
		err=p.updateRecords()
		if err!=nil {
			return err
		}

		p.Tactics.Update()

		PrintAnalysisInfo(p)

		//
		curServerTime:=time.Now().Add(p.serverErrorTime*time.Second)
		curServerTime=curServerTime.Add(getOptFreTimeDur(p.CurTP.OptFrequency))

		log.Infof("预计下一次操作时间：%s ... ...",curServerTime.Format("2006-01-02 15:04:05"))

	}

	//


	//
	return nil
}
func (p *BinanceEx) ping() error {
	p.lastPingTime=time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()
	//
	err:= p.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("Binance ping failed, err: %v", err)
	}
	log.Infoln("Binance Ping success")

	return nil
}
func (p *BinanceEx) updateRecords() error{

	var err error
	//elapsed:=time.Since(p.lastGetRTime)
	//if elapsed>=PING_SPACE{
		//看是否需要请求数据

		//
		curServerTime:=time.Now().Add(p.serverErrorTime*time.Second)
		curOptTime:=getOptPreTime(p.CurTP.OptFrequency,curServerTime)


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
			lastLocalRTime=curServerTime
		}
		//
		err=p.getRecords(curServerTime,lastLocalRTime)
		if err!=nil{
			return err
		}
	//}

	return nil
}
func (p *BinanceEx) getRecords(startTime,endTime time.Time) error {


	//

	//
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()

	startT:=startTime.Unix()*1000
	endT:=endTime.Unix()*1000
	if startTime.Equal(endTime) {
		endT=0
		startT=0
		log.Infof("向服务器请求[最开始->%s]的记录", startTime.Format("2006-01-02 15:04:05"))
	}else {
		log.Infof("向服务器请求[%s->%s]的记录",endTime.Format("2006-01-02 15:04:05"),
			startTime.Format("2006-01-02 15:04:05"))
	}

	records, err := p.client.GetRecords(ctx, p.CurTP.GetBase(), p.CurTP.GetQuote(),
		GetOptFreTimeStr(p.CurTP.OptFrequency), startT,endT,0)
	if err != nil {
		return fmt.Errorf("Get %s records failed, err: %v",p.CurTP.Name,err)
	}

	var modify []int
	rLen:=len(records)
	if rLen>0{

		for i,x:=range records{
			if p.curRecords.Add(&x) {
				modify=append(modify, i)
			}
		}

	}
	PrintGetRecordsInfo(records,modify)

	if len(modify)>0 {
		p.curRecords.Sort()
		//本地同步最新数据
		fErr:=SaveData(getDataFileName(p),&p.curRecords)
		if fErr!=nil {
			log.Errorf("保存文件失败：%+v",fErr)
		}
		//
		p.refreshMACD()
	}
	//



	return nil
}
func (p *BinanceEx) getServerTime() error {

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()

	serverTime, err := p.client.GetTime(ctx)
	if err != nil {
		log.Fatalf("Binance get time failed, err: %v", err)
	}
	p.serverErrorTime=serverTime.Sub(time.Now())/time.Second;
	log.Infof("Binance Server time: %v  误差时间：%v秒", serverTime,p.serverErrorTime)

	return err
}
func (p *BinanceEx)refreshMACD(){
	rLen:=len(p.curRecords.Records)
	if rLen<=30 {
		return
	}
	closeRs:=p.curRecords.GetCloseRecords()

	p.curMACDs=MACDDatas{}
	outMACD, outMACDSignal, outMACDHist := talib.MacdFix(closeRs, 9)

	for  i,x:=range outMACD{
		macd:=MACDData{x,outMACDSignal[i],outMACDHist[i],i}
		p.curMACDs=append(p.curMACDs,macd)
	}

	PrintMACDsInfo(p.curMACDs,p.curRecords.Records,SHOW_MAX_NUM)
	//

}
func (p *BinanceEx)Execute(cmd OptRecord) {
	curRecord:=p.curRecords.GetLastRecord()
	lastRecord:=p.getLastRecord()
	if curRecord==nil||lastRecord==nil {
		log.Errorf("下单失败：找不到数据")
		return
	}
	cmd.indexOpt=p.curOptIndex
	switch cmd.optType {
	case binance.OrderBuy:
		log.Infof("\n建议您买入 ^_^    原因:%s",reasonStr[cmd.reason])
		//
		//定价
		curAPrice:=curRecord.Close
		if  curRecord.Close>curRecord.Open{
			curAPrice=getAverage(lastRecord)
		}
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

		if free>p.CurTP.MinOrderTotalPrice&&free>p.CurTP.MinTradeNum*curAPrice {
			canBuyNum=free/(curAPrice*(1+p.Fee))
			canBuyNum=getPrecisionFloat(canBuyNum,p.CurTP.MinTradeNum)
		}
		if canBuyNum<=p.CurTP.MinTradeNum {
			log.Infof("下单失败：您的 %s 不足，下单买入:%f",p.CurTP.GetQuote(),free)
		}else {

			cmd.time=time.Now()

			amount:=canBuyNum
			cost:=amount*curAPrice+amount*curAPrice*p.Fee

			//

			free-=cost
			if free<0 {
				return
			}

			//
			cmd.orderID=p.trade(binance.OrderSide.String(binance.OrderBuy),amount,curAPrice)
			//
			p.opts=append(p.opts,cmd)
			if cmd.orderID>0 {
				log.Infof("成功下单：买 orderID=%d amount=%f  price=%f",cmd.orderID,amount,curAPrice)
			}
		}
	case binance.OrderSell:
		log.Infof("\n建议您买出 ^_^   原因:%s",reasonStr[cmd.reason])
		//定价
		curAPrice:=curRecord.Close
		curAPrice=getPrecisionFloat(curAPrice,p.CurTP.MinTradePrice)
		if curAPrice<=p.CurTP.MinTradePrice {
			return
		}

		ownBaseFree:=0.0

		ownBase:=p.account.getBalance(p.CurTP.GetBase())
		if ownBase!=nil {
			//调试用-----
			if p.CurTP.OptFrequency==Hour {
				ownBaseFree=ownBase.Free-19.7
			}else {
				ownBaseFree=ownBase.Free
			}
			//----------
		}
		income:=ownBaseFree-ownBaseFree*p.Fee
		if income>0&&income>p.CurTP.MinTradeNum{
			realIncome:=getPrecisionFloat(income,p.CurTP.MinTradeNum)
			sellMoney:=realIncome*curAPrice
			if realIncome>p.CurTP.MinTradeNum&&sellMoney>p.CurTP.MinOrderTotalPrice {
				cmd.time=time.Now()
				cmd.orderID=p.trade(binance.OrderSide.String(binance.OrderSell),realIncome,curAPrice)
				//
				p.opts=append(p.opts,cmd)
				//
				if cmd.orderID>0 {
					log.Infof("成功下单：卖 orderID=%d amount=%f  price=%f",cmd.orderID,realIncome,curAPrice)
				}
			} else {
				log.Errorln("下单失败：仓位太低")
			}
		}



	default:
		log.Errorln("错误")
	}
}
func (p *BinanceEx)getTicker() error{
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()
	//
	ticker, err := p.client.GetTicker(ctx, p.CurTP.GetBase(), p.CurTP.GetQuote())
	if err != nil {
		log.Fatalf("Binance 获得 ticker failed, err: %v", err)
	}

	log.Infof("%s Ticker: %+v",p.CurTP.Name,ticker)

	return err
}
func (p *BinanceEx)getAccount() error{

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()
	//账户信息
	b, err := p.client.GetAccount(ctx, 0)
	if err!=nil {
		log.Fatalf("Binance get Account, err: %v", err)
	}
	exce:=[]string{p.CurTP.GetQuote(),p.CurTP.GetBase()}
	p.account.balances=GetNonZeroBalance(b,exce)
	log.Infof("账户情况: %+v",p.account.balances)

	if len(p.initAssets.Currency)<=0 {
		balance:=p.account.getBalance(p.CurTP.GetQuote())
		balanceB:=p.account.getBalance(p.CurTP.GetBase())
		if balance!=nil&&balanceB!=nil {
			p.initAssets=model.Balance(*balance)
			p.initBase=model.Balance(*balanceB)
		}else{
			p.initAssets=model.Balance{p.CurTP.GetQuote(),0,0}
			p.initBase=model.Balance{p.CurTP.GetBase(),0,0}
		}
		p.initExRate=getUSDToCNY()
		log.Infof("当前人民币汇率：$1 美元=￥%f 人民币",p.initExRate)
	}
	return err
}
func (p *BinanceEx)updateOrders() bool{
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()

/*	orders, err := p.client.GetOrders(ctx,p.CurTP.GetBase(),p.CurTP.GetQuote(), 0)
	if err != nil {
		log.Fatalf("获取open订单失败, err: %v", err)
	}

	log.Infof("获取open订单成功: %+v", orders)*/

	orders, err := p.client.GetAllOrders(ctx,p.CurTP.GetBase(),p.CurTP.GetQuote(), 0,10,0)
	if err != nil {
		log.Fatalf("获取订单失败, err: %v", err)
	}
	if len(orders)==0 {
		return false
	}
	//过滤掉以前的
	var realOrders []model.Order
	for _,x:=range orders{
		if x.Time.Equal(p.StartTime)||x.Time.After(p.StartTime){
			realOrders=append(realOrders,x)
		}
	}
	if len(realOrders)==0 {
		return false
	}
	//log.Infof("获取订单成功: %+v", orders)
	//
	p.account.refreshOrders(realOrders)
	//处理滑单的
	openOrders:=p.account.getOpenOrders()

	if len(openOrders)>0 {
		cancelNum:=0
		for _, y := range openOrders {
			x:=&p.account.orders[y]
			/*opt := p.getOptRecordByOrderID(x.ID)
			if opt == nil {
				continue
			}*/
			//
			p.cancelOrder(x)
			cancelNum++
		}
		if cancelNum>0 {
			return true
		}
	}
	return false
}
func (p *BinanceEx)getOptRecordByOrderID(ID int64) *OptRecord {
	for _,x:=range p.opts{
		if x.orderID==ID{
			return &x
		}
	}
	return nil
}
func (p *BinanceEx)cancelOrder(order *model.Order)  {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()

	err:=p.client.CancelOrder(ctx,p.CurTP.GetBase(),p.CurTP.GetQuote(),order.ID,0)
	if err != nil {
		log.Fatalf("取消订单失败, err: %v", err)
	}

	log.Infof("取消订单成功:%+v",*order)
}
func (p *BinanceEx)getMyTrades()  {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()

	trades, err := p.client.GetMyTrades(ctx,p.CurTP.GetBase(),p.CurTP.GetQuote(), 0,10,0)
	if err != nil {
		log.Fatalf("获取交易记录失败, err: %+v", err)
	}
	if len(trades)==0 {
		return
	}
	//过滤掉以前的
	var realTrades []model.Trade
	for _,x:=range trades{
		if x.Time.Equal(p.StartTime)||x.Time.After(p.StartTime){
			realTrades=append(realTrades,x)
		}
	}
	if len(realTrades)==0 {
		return
	}
	p.account.refreshTrades(realTrades)

}
func (p *BinanceEx)getLastRecord() *model.Record{

	curServerTime:=time.Now().Add(p.serverErrorTime*time.Second)
	curServerTime=curServerTime.Add(-getOptFreTimeDur(p.CurTP.OptFrequency))

	curOptTime:=getOptPreTime(p.CurTP.OptFrequency,curServerTime)

	//
	return p.curRecords.getRecordByTime(curOptTime)
}
func (p *BinanceEx)trade(orderSide string,quantity float64, price float64) int64{
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_TIME)
	defer cancel()

	orderId,err:=p.client.Trade(ctx,p.CurTP.GetBase(),p.CurTP.GetQuote(),
		orderSide,binance.OrderType.String(binance.OrderLimit),
			binance.TimeInForce.String(binance.GTC),quantity,price,0,0,0)
	if err != nil {
		log.Fatalf("下单失败, err: %+v", err)
	}else if orderId==0 {
		log.Errorf("服务器返回订单为空：why?")
	}


	return orderId
}