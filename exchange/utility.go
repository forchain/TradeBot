package exchange

import (
	"github.com/forchain/cryptotrader/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"github.com/forchain/cryptotrader/binance"
	"bytes"
	"strings"
	"net/http"
	"strconv"
	"math"
	"os"
	"path/filepath"
)

func getAverage(record *model.Record) float64  {
	return (record.Open+record.Close+record.High+record.Low)/4.0
}
func getGeometryAverage(record *model.Record) float64  {
	return math.Pow(record.Open*record.Close*record.High*record.Low,1.0/4.0)
}
func getHourTime(t time.Time) time.Time{
	return time.Date(t.Year(),t.Month(),t.Day(), t.Hour(),0,0,0,time.Local)
}
func GetNonZeroBalance(balances []model.Balance,exception []string) []model.Balance {
	res := []model.Balance{}
	for _, v := range balances {
		isE:=false
		for _,x:=range exception{
			if v.Currency==x {
				isE=true
			}
		}
		if isE||v.Free != 0 || v.Frozen != 0 {
			res = append(res, v)
		}
	}

	return res
}
/*获得当前的人民币汇率*/
func getUSDToCNY()float64{
	resp, err := http.Get("https://finance.google.cn/finance/converter?a=1&from=USD&to=CNY")
	if err != nil {
		// handle error
		log.Println(err)
		return 0
	}

	defer resp.Body.Close()

	buf := bytes.NewBuffer(make([]byte, 0, 512))

	//length, _ := buf.ReadFrom(resp.Body)
	buf.ReadFrom(resp.Body)
	bufStr:=string(buf.Bytes())
	//提取里面的某个文字
	index0:=strings.Index(bufStr, "bld>")
	index1:=strings.Index(bufStr,"CNY</span>")
	resultStr:=bufStr[index0+4:index1]
	resultStr=strings.Trim(resultStr," ")
	cny,err:=strconv.ParseFloat(resultStr,64)
	if err != nil {
		// handle error
		log.Println(err)
		return 0
	}
	return cny
}
func getPrecisionFloat(src,pre float64)float64{
	//

	preStr:=fmt.Sprintf("%f",pre)
	t:=strings.Index(preStr,".")
	dNum:=1
	if t!=-1 {
		t0:=preStr[t+1:]
		tLen:=len(t0)
		if tLen>0 {
			for {
				tLen:=len(t0)
				if tLen<=0 {
					break
				}
				if t0[tLen-1]!='0' {
					dNum = len(t0)
					break
				}else {
					t0=t0[0:tLen-1]
				}
			}
		}
	}else{
		dNum=1
	}
	//
	srcTemp:=math.Trunc(src/pre)
	srcTemp*=pre
	d10:=math.Pow(10,float64(dNum))
	result:=math.Trunc(srcTemp*d10)/d10

	return result
}
func GetDataFileList(path string,match []string)[]os.FileInfo {
	var result []os.FileInfo
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if ( f == nil ) {return err}
		if !f.IsDir()&&len(f.Name())>0{
			for _,v:=range match{
				if !strings.Contains(f.Name(), v)  {
					return nil
				}
			}
			result=append(result,f)
		}
		return nil
	})
	if err != nil {
		return []os.FileInfo{}
	}
	return result
}

//-----------------------------------------------------------
func PrintGetRecordsInfo(records []model.Record,modify []int){

	modifyLen:=len(modify)
	rLen:=len(records)
	if rLen<=0{
		log.Infof("本次没有获得任何记录")
	}else {
		showStr:=fmt.Sprintf("\n本次获得记录总数：%d条 时间跨度：[%s->%s]",rLen,
			records[0].Time.Format("2006-01-02 15:04:05"),
			records[rLen-1].Time.Format("2006-01-02 15:04:05"))

		showStr+=fmt.Sprintf("  本地更新数：%d条\n",modifyLen)
		var i int
		for i=0;i< 123;i++  {
			showStr+="-"
		}

		showStr+="\n|    ID  |"
		showStr+="     开盘价     |"
		showStr+="     收盘价     |"
		showStr+="     最高价     |"
		showStr+="     最低价     |"
		showStr+="     交易量     |"
		showStr+="          时间        |  状态  |\n"
		for i=0;i< 123;i++  {
			showStr+="-"
		}
		showStr+="\n"

		rs:=RSort(records)

		for i,x:=range rs{
			var hasModify=false
			for _,y:=range modify{
				if (rLen-1-y)==i {
					hasModify=true
					break
				}
			}
			stateStr:="   **   "
			if hasModify {
				stateStr="  更新  "
			}
			showStr+=fmt.Sprintf("%8d%16g%16g%16g%16g%16g   %20s  %s\n",i,x.Open,x.Close,x.High,x.Low,x.Vol,
				x.Time.Format("2006-01-02 15:04:05"),stateStr)
		}
		log.Info(showStr)
	}
}

func PrintLocalRecordsInfo(records []model.Record,showNum int){

	rLen:=len(records)
	if rLen<=0{
		log.Infof("\n本地没有任何交易记录")
	}else {
		showStr:=fmt.Sprintf("\n本地读取记录总数：%d条 时间跨度：[%s->%s]",rLen,
			records[0].Time.Format("2006-01-02 15:04:05"),
			records[rLen-1].Time.Format("2006-01-02 15:04:05"))

		rShowNum:=showNum
		if rLen<showNum {
			rShowNum=rLen
		}
		showStr+=fmt.Sprintf("  只显示：%d条\n",rShowNum)
		var i int
		for i=0;i< 123;i++  {
			showStr+="-"
		}

		showStr+="\n|    ID  |"
		showStr+="     开盘价     |"
		showStr+="     收盘价     |"
		showStr+="     最高价     |"
		showStr+="     最低价     |"
		showStr+="     交易量     |"
		showStr+="          时间        |   状态    |\n"
		for i=0;i< 123;i++  {
			showStr+="-"
		}
		showStr+="\n"

		//
		rs:=RSort(records)

		var x model.Record
		for i,x=range rs{
			if i>=showNum {
				break
			}
			showStr+=fmt.Sprintf("%8d%16g%16g%16g%16g%16g   %20s\n",i,x.Open,x.Close,x.High,x.Low,x.Vol,
				x.Time.Format("2006-01-02 15:04:05"))
		}


		if rShowNum<rLen {
			if rLen-rShowNum-1>0 {
				showStr+=fmt.Sprintf("\n\n.......省略[%d]条记录........\n\n",rLen-rShowNum-1)
			}


			showStr+="..........\n"
			i=rLen-1
			x=rs[i]
			showStr+=fmt.Sprintf("%8d%16g%16g%16g%16g%16g   %20s\n",i,x.Open,x.Close,x.High,x.Low,x.Vol,
				x.Time.Format("2006-01-02 15:04:05"))
		}
		log.Info(showStr)
	}
}

func PrintMACDsInfo(macds MACDDatas,records []model.Record,showNum int){
	mLen:=len(macds)
	if mLen<=0 {
		return
	}
	rLen:=len(records)
	if rLen<=0{
		log.Infof("\n没有任何MACD数据")
	}else {
		showStr:=fmt.Sprintf("\nMACD总数：%d条 时间跨度：[%s->%s]",rLen,
			records[0].Time.Format("2006-01-02 15:04:05"),
			records[rLen-1].Time.Format("2006-01-02 15:04:05"))

		rShowNum:=showNum
		if rLen<showNum {
			rShowNum=rLen
		}
		showStr+=fmt.Sprintf("  只显示：%d条\n",rShowNum)
		var i int
		for i=0;i< 80;i++  {
			showStr+="-"
		}

		showStr+="\n|    ID   |"
		showStr+="      DIF      |"
		showStr+="      DEA      |"
		showStr+="      MACD     |"
		showStr+="         时间        |\n"
		for i=0;i< 80;i++  {
			showStr+="-"
		}
		showStr+="\n"

		//
		for i=rLen-1;i>=0;i--  {
			if (rLen-i)>showNum {
				break
			}
			showStr+=fmt.Sprintf("%8d%16.2f%16.2f%16.2f   %20s\n",rLen-i-1,macds[i].DIF,macds[i].DEA,
				macds[i].MACD,records[i].Time.Format("2006-01-02 15:04:05"))
		}
		if rShowNum<rLen {
			showStr+=fmt.Sprintf("\n.......后面还有[%d]条记录........\n",rLen-rShowNum)
		}
		log.Info(showStr)
	}
}


func PrintOptimalMACDsInfo(macds MACDDatas,records []model.Record,isDetailed bool){

	mLen:=len(macds)
	if mLen>0{
		showStr:=fmt.Sprintf("\n正在进行MACD降噪处理...\nMACD优化总数：%d条\n",mLen)
		if !isDetailed {
			log.Info(showStr)
			return
		}
		var i int
		for i=0;i< 80;i++  {
			showStr+="-"
		}

		showStr+="\n|    ID   |"
		showStr+="      DIF      |"
		showStr+="      DEA      |"
		showStr+="      MACD     |"
		showStr+="         时间        |\n"
		for i=0;i< 80;i++  {
			showStr+="-"
		}
		showStr+="\n"

		//
		for i=mLen-1;i>=0;i--  {
			showStr+=fmt.Sprintf("%8d%16.2f%16.2f%16.2f   %20s\n",mLen-i-1,macds[i].DIF,macds[i].DEA,
				macds[i].MACD,records[macds[i].index].Time.Format("2006-01-02 15:04:05"))
		}
		log.Info(showStr)
	}
}
func PrintDebugInfo(p *BinanceDebugEx){

	showStr:=fmt.Sprintf("\n\n******%s调试版：测试自动化交易 %s 分析报告*****\n",p.GetExchange().Name,p.GetExchange().CurTP.Name)
	var optName string
	var i int

	oLen:=len(p.opts)
	if oLen>0 {

		showStr+="\n操作记录：\n"
		for i=0;i< 80;i++  {
			showStr+="-"
		}
		showStr+="\n| 编号 |"
		showStr+="   ID  |"
		showStr+="  操作  |"
		showStr+="    订单号    |"
		showStr+="         时间        |"
		showStr+="         原因        |\n"
		for i=0;i< 80;i++  {
			showStr+="-"
		}
		showStr+="\n"
		//

		for i,x:=range p.opts{
			if x.optType==binance.OrderBuy {
				optName =" 买入 "
			}else {
				optName=" 卖出 "
			}
			tempStr:=reasonStr[x.reason]
			showStr+=fmt.Sprintf("%6d%6d%7s%16d%22s%18s\n",i,x.indexOpt,optName,x.orderID,
				x.time.Format("2006-01-02 15:04:05"), tempStr)
		}

		for i=0;i< 80;i++  {
			showStr+="-"
		}
	}
	//
	rLen:=len(p.account.orders)
	if rLen>0 {

		showStr+="\n\n\n订单情况：\n"

		for i=0;i< 83;i++  {
			showStr+="-"
		}
		showStr+="\n|     订单号     |"
		showStr+="  操作  |"
		showStr+="      状态      |"
		showStr+="        数目       |"
		showStr+="          价格        |\n"
		for i=0;i< 83;i++  {
			showStr+="-"
		}
		showStr+="\n"
		//

		for _,x:=range p.account.orders{

			showStr+=fmt.Sprintf("%16d%8s%16s%20g%20.16g\n",x.ID,x.Side,x.Status,x.Amount,x.Price)
		}
		for i=0;i< 83;i++  {
			showStr+="-"
		}
	}
	//
	tLen:=len(p.account.trades)
	if tLen>0 {

		showStr+="\n\n\n交易记录：\n"

		for i=0;i< 187;i++  {
			showStr+="-"
		}
		showStr+="\n|     交易号     |"
		showStr+="  类型  |"
		showStr+="      交易数目      |"
		showStr+="      成交价格      |"
		showStr+="      手续费用      |"
		showStr+="      持仓数目       |"
		showStr+="      持仓市值       |"
		showStr+="      账户余额       |"
		showStr+="        盈亏        |"
		showStr+="         时间        |\n"

		for i=0;i< 187;i++  {
			showStr+="-"
		}
		showStr+="\n"
		//
		firstA:=p.initAssets.Free+p.initAssets.Frozen

		var totalFee float64=0
		balance:=p.account.getBalance(p.CurTP.GetQuote())
		for _,x:=range p.account.trades{
			ownB:=p.account.getOwnBaseByTime(x.Time)
			baV:=p.account.getBalanceByTime(x.Time)+firstA+ownB*x.Price
			/*baV,err:=strconv.ParseFloat(x.Raw,64)
			if err!=nil {
				baV=0
			}*/
			if x.Type==TradeType.String(TradeBuy) {
				totalFee+=x.Commission
			}else {
				totalFee+=x.Commission*x.Price
			}
			showStr+=fmt.Sprintf("%16d%7s%20g%20.16g%20g%20g%20.16g%20.10g%20.10g%22s\n",x.ID,x.Type,x.Amount,x.Price,
				x.Amount*x.Price*p.Fee,ownB,ownB*x.Price,baV,baV-firstA,
				x.Time.Format("2006-01-02 15:04:05"))
		}


		free:=balance.Free+balance.Frozen+p.account.getOwnBase()*p.curRecords.GetLastRecord().Close

		showStr+=fmt.Sprintf("\n结束\n%s:账户总资产：%f--->%s账户总资产:%f   单位[%s]\n资产总收益:%f     累计手续费:%f",
			p.curRecords.GetFirstRecord().Time.Format("2006-01-02 15:04:05"),firstA,
			p.curRecords.GetLastRecord().Time.Format("2006-01-02 15:04:05"),free,p.CurTP.GetQuote(),
			free-firstA,totalFee)

		huiPre:=(free-firstA)/firstA*100
		//
		firstBase:=firstA/p.account.trades[0].Price
		endBase:=(balance.Free+balance.Frozen)/p.curRecords.GetLastRecord().Close+p.account.getOwnBase()


		handerHui:=(p.curRecords.GetLastRecord().Close-p.curRecords.GetFirstRecord().Close)/p.curRecords.GetFirstRecord().Close*100
		//
		showStr+=fmt.Sprintf("\n投资回报率:%.2f%%   持币分析：[%f]->[%f]=%f 个%s[%.2f%%]   币价分析：[%f]->[%f]=%f %s[%.2f%%]",huiPre,
			firstBase,endBase,endBase-firstBase,p.CurTP.GetBase(),(endBase-firstBase)/firstBase*100,
			p.curRecords.GetFirstRecord().Close,p.curRecords.GetLastRecord().Close,
			p.curRecords.GetLastRecord().Close-p.curRecords.GetFirstRecord().Close,
				p.CurTP.GetQuote(),handerHui)

		//
		var totalOptTrade int=1
		for i:=0;i<tLen ;i++  {

			if i>0&&p.account.trades[i].Type!=p.account.trades[i-1].Type  {
				totalOptTrade++
			}
		}
		optTradeFrequency:=float64(totalOptTrade)/(float64(p.curRecords.GetLastRecord().Time.Sub(p.curRecords.GetFirstRecord().Time))/float64(getOptFreTimeDur(p.CurTP.OptFrequency)))
		//

		showStr+=fmt.Sprintf("\n与一直持有情况下比，投资回报率:%.2f%%     总有效操作数：%d   周期操作频率：%5.2f/%s",huiPre-handerHui,totalOptTrade,optTradeFrequency,GetOptFreTimeStr(p.CurTP.OptFrequency))

		//
		if huiPre<0 {
			showStr+="\n\n悲剧啊 骚年  亏本了  赶紧改*******************\n"
		}else {
			showStr+="\n\n恭喜你 赚钱了"
		}

		if endBase>firstBase {
			showStr+="还好，你的币是一直在增加的O(∩_∩)O\n"
		}else{
			showStr+="晕死，币在减少\n"
		}
		showStr+="\n"
	}



	//

	//


	log.Info(showStr)
	showStr=fmt.Sprintf("初始账户情况：%+v\n\n",p.initAssets)+showStr
	filePath:=p.LogDir+p.CurTP.Name+"_"+GetOptFreTimeStr(p.CurTP.OptFrequency)+".report"
	saveReportToFile(showStr,filePath)
}

func PrintAnalysisInfo(p *BinanceEx){

	oLen:=len(p.opts)
	if oLen<=0 {
		return
	}

	showStr:=fmt.Sprintf("\n\n******%s自动化交易 %s 分析报告*****\n",p.GetExchange().Name,p.GetExchange().CurTP.Name)
	var optName string
	var i int


	if oLen>0 {

		showStr+="\n操作记录：\n"
		for i=0;i< 83;i++  {
			showStr+="-"
		}
		showStr+="\n| 编号 |"
		showStr+="   ID  |"
		showStr+="  操作  |"
		showStr+="    订单号    |"
		showStr+="         时间        |"
		showStr+="         原因        |\n"
		for i=0;i< 83;i++  {
			showStr+="-"
		}
		showStr+="\n"

		//

		for i,x:=range p.opts{
			if x.optType==binance.OrderBuy {
				optName =" 买入 "
			}else {
				optName=" 卖出 "
			}
			tempStr:=reasonStr[x.reason]
			showStr+=fmt.Sprintf("%6d%6d%7s%16d%22s%18s\n",i,x.indexOpt,optName,x.orderID,
				x.time.Format("2006-01-02 15:04:05"), tempStr)
		}

		for i=0;i< 83;i++  {
			showStr+="-"
		}
	}
	//
	rLen:=len(p.account.orders)
	if rLen>0 {

		//
		showStr+="\n\n\n订单情况：\n"

		for i=0;i< 83;i++  {
			showStr+="-"
		}
		showStr+="\n|     订单号     |"
		showStr+="  操作  |"
		showStr+="      状态      |"
		showStr+="        数目       |"
		showStr+="          价格        |\n"
		for i=0;i< 83;i++  {
			showStr+="-"
		}
		showStr+="\n"
		//



		for _,x:=range p.account.orders{

			showStr+=fmt.Sprintf("%16d%8s%16s%20g%20.16g\n",x.ID,x.Side,x.Status,x.Amount,x.Price)
		}
		for i=0;i< 83;i++  {
			showStr+="-"
		}
	}
	//
	tLen:=len(p.account.trades)
	if tLen>0 {

		showStr+="\n\n\n交易记录：\n"

		for i=0;i< 186;i++  {
			showStr+="-"
		}
		showStr+="\n|     交易号     |"
		showStr+="  类型  |"
		showStr+="      交易数目      |"
		showStr+="      成交价格      |"
		showStr+="      手续费用      |"
		showStr+="      持仓数目       |"
		showStr+="      持仓市值       |"
		showStr+="      账户余额       |"
		showStr+="        盈亏        |"
		showStr+="         时间        |\n"

		for i=0;i< 186;i++  {
			showStr+="-"
		}
		showStr+="\n"
		//
		startOptTime:=getOptPreTime(p.CurTP.OptFrequency,p.StartTime)
		startOptRecord:=p.curRecords.getRecordByTime(startOptTime)

		firstA:=p.initAssets.Free+p.initAssets.Frozen
		firstB:=p.initBase.Free+p.initBase.Frozen
		firstBMoney:=startOptRecord.Close*firstB

		var totalFee float64=0
		balance:=p.account.getBalance(p.CurTP.GetQuote())
		for _,x:=range p.account.trades{
			ownB:=p.account.getOwnBaseByTime(x.Time)+firstB
			baV:=p.account.getBalanceByTime(x.Time)+firstA+ownB*x.Price+firstBMoney
			if x.Type==TradeType.String(TradeBuy) {
				totalFee+=x.Commission
			}else {
				totalFee+=x.Commission*x.Price
			}

			showStr+=fmt.Sprintf("%16d%7s%20g%20.16g%20g%20g%20.16g%20.10g%20.10g%22s\n",x.ID,x.Type,x.Amount,x.Price,
				x.Commission,ownB,ownB*x.Price,baV,baV-firstA-firstBMoney,
				x.Time.Format("2006-01-02 15:04:05"))

		}
		ownBaseAll:=0.0
		baseBa:=p.account.getBalance(p.CurTP.GetBase())
		if baseBa!=nil {
			ownBaseAll=baseBa.Free+baseBa.Frozen
		}
		free:=balance.Free+balance.Frozen+ownBaseAll*p.curRecords.GetLastRecord().Close

		showStr+=fmt.Sprintf("\n结束\n%s:账户总资产：%f--->%s账户总资产:%f   单位[%s]\n资产总收益:%f     累计手续费:%f",
			p.StartTime.Format("2006-01-02 15:04:05"),firstA+firstBMoney,
			p.curRecords.GetLastRecord().Time.Format("2006-01-02 15:04:05"),free,p.CurTP.GetQuote(),
			free-firstA-firstBMoney,totalFee)

		huiPre:=(free-firstA-firstBMoney)/(firstA+firstBMoney)*100
		//
		firstBase:=firstA/startOptRecord.Close+firstB
		endBase:=(balance.Free+balance.Frozen)/p.curRecords.GetLastRecord().Close+ownBaseAll

		//

		handerHui:=(p.curRecords.GetLastRecord().Close-startOptRecord.Close)/startOptRecord.Close*100
		//
		showStr+=fmt.Sprintf("\n投资回报率:%.2f%%   持币分析：[%f]->[%f]=%f 个%s[%.2f%%]   币价分析：[%f]->[%f]=%f %s[%.2f%%]",huiPre,
			firstBase,endBase,endBase-firstBase,p.CurTP.GetBase(),(endBase-firstBase)/firstBase*100,
			startOptRecord.Close,p.curRecords.GetLastRecord().Close,
			p.curRecords.GetLastRecord().Close-startOptRecord.Close,
			p.CurTP.GetQuote(),handerHui)
		//
		var totalOptTrade int=1
		for i:=0;i<tLen ;i++  {

			if i>0&&p.account.trades[i].Type!=p.account.trades[i-1].Type  {
				totalOptTrade++
			}
		}
		optTradeFrequency:=float64(totalOptTrade)/(float64(p.curRecords.GetLastRecord().Time.Sub(p.StartTime))/float64(getOptFreTimeDur(p.CurTP.OptFrequency)))
		//
		showStr+=fmt.Sprintf("\n与一直持有情况下比，投资回报率:%.2f%%     总有效操作数：%d   周期操作频率：%5.2f/%s",huiPre-handerHui,totalOptTrade,optTradeFrequency,GetOptFreTimeStr(p.CurTP.OptFrequency))
		//
		exRate:=getUSDToCNY()
		showStr+=fmt.Sprintf("\n当前人民币汇率：$1 美元=￥%f 人民币   开始汇率[%f]  升值[￥%f]  总共[￥%f]\n",
			exRate,p.initExRate,exRate-p.initExRate,(exRate-p.initExRate)*firstA)
		if huiPre<0 {
			showStr+="\n\n悲剧啊 骚年  亏本了  赶紧改*******************\n"
		}else {
			showStr+="\n\n恭喜你 赚钱了"
		}

		if endBase>firstBase {
			showStr+="还好，你的币是一直在增加的O(∩_∩)O\n"
		}else{
			showStr+="晕死，币在减少\n"
		}
		showStr+="\n"
	}



	//

	//


	log.Info(showStr)
	showStr=fmt.Sprintf("初始账户情况：%+v %+v\n\n",p.initBase,p.initAssets)+showStr
	filePath:=p.LogDir+p.CurTP.Name+"_"+GetOptFreTimeStr(p.CurTP.OptFrequency)+".report"
	saveReportToFile(showStr,filePath)
}
func saveReportToFile(msg,fileName string){

	f,err:= os.Create(fileName)

	if err != nil {
		log.Infof("保存报告文件失败:%s",fileName)
		return
	}
	defer func() {
		err=f.Close()
	}()

	log.Infof("生成简报O(∩_∩)O:%s",fileName)
	//
	f.WriteString(msg)
	//

}
/*调试开始的时间点*/
var DebugStartTime *time.Time
var DebugEndTime *time.Time