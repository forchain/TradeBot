package exchange

import (

	"strings"
	"github.com/forchain/cryptotrader/model"
	"os"
	"encoding/json"
	"sort"
	"time"
	"github.com/forchain/cryptotrader/binance"
	log "github.com/sirupsen/logrus"
	"fmt"
	"strconv"
)

/*交易所*/
const (
	Binance="Binance"
	OK="OK"
	ZB="ZB"
)
/*数据文件中记录的最大条数*/
const MaxRecordItemNum=500

type IExchange interface {
	Init()
	Exit()
	Update() error
	GetExchange() (*Exchange)
	Execute(cmd OptRecord)
}
type Exchange struct {
	Name string
	Fee float64
	APIKey string
	APISecretKey string
	LogDir string
	StartTime time.Time
	CurTP TradePairConfig
	Tactics ITactics
	StopLoss float64
	StopGain float64
}
type ExchangeRecords struct {
	Records []model.Record
}
func (p *ExchangeRecords)Has(r *model.Record) bool{
	for _,x:=range p.Records{
		if r.Time.Equal(x.Time) {
			return true
		}
	}
	return false
}
func (p *ExchangeRecords)Add(r *model.Record) bool{
	if p.Has(r) {
		return false
	}
	p.Records=append(p.Records, *r)
	return true
}
func (p *ExchangeRecords)GetFirstRecord() *model.Record{
	rLen:=len(p.Records)
	if rLen<=0 {
		return nil
	}

	return &p.Records[0]
}
func (p *ExchangeRecords)GetLastRecord() *model.Record{
	rLen:=len(p.Records)
	if rLen<=0 {
		return nil
	}

	return &p.Records[rLen-1]
}
func (p *ExchangeRecords)GetCloseRecords() []float64{
	var result []float64
	for _,x:=range p.Records{
		result=append(result,x.Close)
	}
	return result
}
func (p *ExchangeRecords)getRecordByTime(time time.Time) *model.Record{
	rLen:=len(p.Records)
	if rLen<=0 {
		return nil
	}
	for _,x:=range p.Records{
		if x.Time.Equal(time) {
			return &x
		}
	}
	return nil
}
func (p *ExchangeRecords)getRecordIndexByTime(time time.Time) int{
	rLen:=len(p.Records)
	if rLen<=0 {
		return -1
	}
	for i,x:=range p.Records{
		if x.Time.Equal(time) {
			return i
		}
	}
	return -1
}
func (p *ExchangeRecords)getRecordBeforeTime(time time.Time) []model.Record{

	var before []model.Record

	for _,x:=range p.Records{
		if x.Time.Equal(time)||x.Time.Before(time) {
			before=append(before,x)
		}
	}
	return before
}
func (p *ExchangeRecords)Sort(){
	rs:=RecordsSort(p.Records)
	sort.Sort(rs)
}
func (p *ExchangeRecords)Join(other *ExchangeRecords){

	for _,v:=range other.Records{
		p.Records=append(p.Records,v)
	}
}


type RecordsSort []model.Record
func (p RecordsSort) Len() int {
	return len(p)
}
func (p RecordsSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p RecordsSort) Less(i, j int) bool {
	return p[i].Time.Before(p[j].Time)
}
type RecordsRSort []model.Record
func (p RecordsRSort) Len() int {
	return len(p)
}
func (p RecordsRSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p RecordsRSort) Less(i, j int) bool {
	return p[i].Time.After(p[j].Time)
}
func RSort(records []model.Record) RecordsRSort{
	var rs RecordsRSort
	for _,xx:=range records{
		rs=append(rs,xx)
	}
	sort.Sort(rs)
	return rs
}


const (
	BTC_USDT="BTC-USDT"
	ETH_USDT="ETH-USDT"
	ETH_BTC="ETH-BTC"
	BNB_USDT="BNB-USDT"
)
const(
	Min5=iota
	Min30
	Hour
	Day
	Week
	Month
)
func getOptFreTimeDur(optFre int) time.Duration{
	switch optFre {
	case Min5:
		return time.Minute*5
	case Min30:
		return time.Minute*30
	case Hour:
		return time.Hour
	default:
		return 0
	}
	return 0
}
func getOptPreTime(optFre int,t time.Time) time.Time {
	switch optFre {
	case Min5:
		minNum:=int(t.Minute()/5)
		minNum*=5
		return time.Date(t.Year(),t.Month(),t.Day(),
			t.Hour(),minNum,0,0,time.Local)
	case Min30:
		if time.Duration(t.Minute())>=time.Minute*30 {
			return time.Date(t.Year(),t.Month(),t.Day(),
				t.Hour(),int(time.Minute*30),0,0,time.Local)
		}else {
			return time.Date(t.Year(),t.Month(),t.Day(),
				t.Hour(),0,0,0,time.Local)
		}

	case Hour:
		return time.Date(t.Year(),t.Month(),t.Day(),
			t.Hour(),0,0,0,time.Local)
	default:
		return t
	}
	return t
}
func GetOptFreTimeStr(optFre int) string{
	switch optFre {
	case Min5:
		return binance.Interval.String(binance.Interval5m)
	case Min30:
		return binance.Interval.String(binance.Interval30m)
	case Hour:
		return binance.Interval.String(binance.Interval1h)
	default:
		return ""
	}
	return ""
}
/*币对的交易配置*/
type TradePairConfig struct {
	Name string
	OptFrequency int
	MinTradeNum float64/*最小交易数*/
	MinOrderTotalPrice float64/*最小下单总金额*/
	MinTradePrice float64/*最小交易价格*/
}
func (p *TradePairConfig) GetBase() string{
	sStr:=strings.Split(p.Name,"-")
	if len(sStr)==2 {
		return sStr[0]
	}
	return ""
}
func (p *TradePairConfig) GetQuote() string{
	sStr:=strings.Split(p.Name,"-")
	if len(sStr)==2 {
		return sStr[1]
	}
	return ""
}

func getDataFileName(exch IExchange) string {

	return "data/"+exch.GetExchange().Name+"/"+GetOptFreTimeStr(exch.GetExchange().CurTP.OptFrequency)+"/"+exch.GetExchange().CurTP.Name+".json"
}
func getDataFilePath(exch IExchange) string {

	return "data/"+exch.GetExchange().Name+"/"+GetOptFreTimeStr(exch.GetExchange().CurTP.OptFrequency)
}
func getRecordsFileInfoSort(dir string,tpName string)RecordsFileInfoSort{
	//获取最近的数据包
	var rfis RecordsFileInfoSort
	fileList:=GetDataFileList(dir,[]string{".json",tpName})
	if len(fileList)<=0 {
		return rfis
	}

	var err error
	var rfi RecordsFileInfo

	for _,v:=range fileList {
		rfi = RecordsFileInfo{File: v}
		err=rfi.Init()
		if err==nil {
			rfis=append(rfis,rfi)
		}
	}
	sort.Sort(rfis)
	return rfis
}
func LoadData(dir string,records *ExchangeRecords,tpName string,isDebug bool) error {
	//获取最近的数据包
	rfis:=getRecordsFileInfoSort(dir,tpName)
	rLen:=len(rfis)
	if rLen<=0 {
		return fmt.Errorf("没有找到任何数据")
	}
	var err error
	var fileName string
	//

	if rLen==1 {
		fileName=dir+"/"+rfis[0].File.Name()
	}else {
		if isDebug {
			var tempRecors ExchangeRecords
			var rtempRecors ExchangeRecords

			for _,r:=range rfis{
				tempRecors=ExchangeRecords{}
				fileName=dir+"/"+r.File.Name()
				curE:=doLoadData(fileName,&tempRecors)

				//过滤
				if DebugStartTime!=nil {
					rtempRecors=ExchangeRecords{}
					for _,tr:=range tempRecors.Records{
						if (tr.Time.After(*DebugStartTime)||tr.Time.Equal(*DebugStartTime))&&
							(tr.Time.Before(*DebugEndTime)||tr.Time.Equal(*DebugEndTime)){
							rtempRecors.Records=append(rtempRecors.Records,tr)
						}
					}
					tempRecors=rtempRecors
				}
				//
				if len(tempRecors.Records)>0 {
					records.Join(&tempRecors)
				}

				if curE!=nil {
					err=curE
				}

			}
			return err
		}else {
			fileName=dir+"/"+rfis[rLen-1].File.Name()
		}

	}

	if fileName=="" {
		return fmt.Errorf("没有找到任何数据")
	}
	return doLoadData(fileName,records)


	//

}
func doLoadData(fileName string,records *ExchangeRecords)error{
	var f *os.File
	var err error

	if IsExist(fileName) {
		f,err= os.Open(fileName)
	} else {
		err=fmt.Errorf("文件不存在")
	}

	if err != nil {
		return err
	}

	defer func() {
		cErr:=f.Close()
		if err==nil {
			err=cErr
		}
	}()
	//
	err=json.NewDecoder(f).Decode(records)
	if err!=nil {
		return err
	}
	log.Infof("读取数据文件：%s",fileName)
	return nil
}

func SaveData(dir string,records *ExchangeRecords,tpName string) error {

	//获取最近的数据包
	rfis:=getRecordsFileInfoSort(dir,tpName)
	sLen:=len(rfis)

	var fileName string
	var err error
	rLen:=len(records.Records)
	if rLen>MaxRecordItemNum {
		var num int=rLen/MaxRecordItemNum+1
		var index int=0
		var tempRecords ExchangeRecords
		var tLen int
		for i:=0; i<num;i++  {
			tempRecords.Records=[]model.Record{}
			for j:=0;j<MaxRecordItemNum ;j++  {
				index=i*MaxRecordItemNum+j
				if index<rLen {
					tempRecords.Records=append(tempRecords.Records,records.Records[index])
				}else {
					break
				}
			}
			tLen=len(tempRecords.Records)
			if tLen>0 {
				if i>0{
					time.Sleep(time.Second*2)
					fileName=dir+"/"+tpName+"_"+strconv.FormatInt(time.Now().Unix(),10)+".json"
				}else {
					if sLen>0 {
						fileName=dir+"/"+rfis[sLen-1].File.Name()
					}else {
						fileName=dir+"/"+tpName+"_"+strconv.FormatInt(time.Now().Unix(),10)+".json"
					}
				}

				erre:=doSaveData(fileName,&tempRecords)
				if err!=nil {
					err=erre
				}
			}else {
				break
			}

		}
	}else {
		if sLen>0 {
			fileName=dir+"/"+rfis[sLen-1].File.Name()
		}else {
			fileName = dir + "/" + tpName + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".json"
		}
		return doSaveData(fileName,records)
	}
	return err
}

func doSaveData(fileName string,records *ExchangeRecords) error {


	//
	var f *os.File
	var err error
	f,err= os.Create(fileName)

	if err != nil {
		return err
	}

	defer func() {
		cErr:=f.Close()
		if err==nil {
			err=cErr
		}
	}()


	//
	encoder:=json.NewEncoder(f)
	encoder.SetIndent("","    ")
	err=encoder.Encode(records)
	//f.Sync()
	log.Infof("保存数据文件：%s",fileName)
	return err
}
/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func IsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

type MACDData struct {
	DIF float64
	DEA float64
	MACD float64
	index int
}
type MACDDatas []MACDData

type Account struct {
	balances []model.Balance
	//canUse float64
	trades []model.Trade
	orders []model.Order
}
func (p *Account)getBalance(name string) *model.Balance{
	for i,x:=range p.balances{
		if x.Currency==name {
			return &p.balances[i]
		}
	}
	return nil
}
func (p *Account)getBalanceByTime(time time.Time) float64{
	var result float64=0
	for _,x:=range p.trades{
		if x.Time.Equal(time)||x.Time.Before(time) {
			if x.Type==TradeType.String(TradeBuy) {
				result-=x.Amount*x.Price
				result-=x.Commission
			}else {
				result+=x.Amount*x.Price
				result-=x.Commission*x.Price
			}

		}

	}
	return result
}
func (p *Account)getOwnBase() float64{
	var result float64=0
	for _,x:=range p.trades{
		if x.Type==TradeType.String(TradeBuy) {
			result+=x.Amount
		}else {
			result-=x.Amount
			result-=x.Commission
		}
	}
	return result
}
func (p *Account)getOwnBaseByTime(time time.Time) float64{
	var result float64=0
	for _,x:=range p.trades{
		if x.Time.Equal(time)||x.Time.Before(time) {
			if x.Type==TradeType.String(TradeBuy) {
				result+=x.Amount
			}else {
				result -= x.Amount
				result-=x.Commission
			}
		}

	}
	return result
}
func (p *Account)getOrders(status int) []int{
	var result []int
	for i,x:=range p.orders{
		if x.Status==OrderStatus.String(status){
			result=append(result,i)
		}
	}
	return result
}
func (p *Account)getOpenOrders() []int{
	var result []int
	for i,x:=range p.orders{
		if x.Status==OrderStatus.String(OrderNew)||
			x.Status==OrderStatus.String(OrderPartiallyFilled){
			result=append(result,i)
		}
	}
	return result
}
func (p *Account)refreshOrders(orders []model.Order){
	for _,x:=range orders{
		has:=false
		for j,y:=range p.orders{
			if x.ID==y.ID{
				has=true
				if  x.Status!=y.Status{
					p.orders[j]=x
					log.Infof("更新order:%+v",p.orders[j])
				}
				break
			}
		}
		if !has {
			p.orders=append(p.orders,x)
			log.Infof("添加order:%+v",x)
		}
	}
}
func (p *Account)refreshTrades(trades []model.Trade){
	for _,x:=range trades{
		has:=false
		for _,y:=range p.trades{
			if x.ID==y.ID{
				has=true
				break
			}
		}
		if !has {
			/*if x.Type==binance.OrderSide.String(binance.OrderSell){
				x.Commission=x.Commission*x.Price
				x.CommissionAsset=x.CommissionAsset+"->原始"
			}*/

			p.trades=append(p.trades,x)
			log.Infof("添加trade:%+v",x)
		}
	}
}
