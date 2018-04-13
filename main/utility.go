package main

import (
	"fmt"
	"os"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/forchain/TradeBot/exchange"
	"time"
)


const FileName  = "data/config.json"
var logFileName string=time.Now().Format("2006-01-02_15-04-05")+".log"

/*交易市场的配置*/
type ExchangeConfig struct {
	Name string
	Fee float64
	APIKey string
	APISecretKey string
	Tactics int
	LogDir string
	IsDebug bool
	CurTP exchange.TradePairConfig
}



type LogFormatter struct {
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {

	temp:=fmt.Sprintf("%s:%s\n",time.Now().Format("2006-01-02 15:04:05"),
		entry.Message)
	if logFile!=nil {
		logFile.WriteString(temp)
	}
	return []byte(temp),nil
}

var ExchangeC ExchangeConfig
var curExchange exchange.IExchange


func Init() bool {
	//
	log.SetFormatter(new(LogFormatter))
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	//
	log.Infoln("欢迎使用自动化交易系统^_^")
	//

	//
	err:=LoadConfig(FileName)
	if err!=nil {
		log.Errorln(err)
		return false
	}
	//
	openLogFile()
	//配置交易策略模块

	curExchange,err=getExchangeInstance()
	if err!=nil {
		log.Errorln(err)
		return false
	}
	//
	log.Infof("当前市场：%s - 交易对：%s  交易费用：%f",curExchange.GetExchange().Name,
		curExchange.GetExchange().CurTP.Name,curExchange.GetExchange().Fee)

	curExchange.Init()
	//

	return true
	//
}
func Exit(){

	curExchange.Exit()

	//err:=SaveConfig(fileName)


	log.Infoln("退出自动化交易系统^_^")

	closeLogFile()

	/*if err!=nil {
		fmt.Println(err)
	}*/
}
func update() bool{

	//time.Sleep(time.Second)
	err:=curExchange.Update()
	if err!=nil {
		log.Errorln(err)
		return false
	}
	//
	return true
}
func Run() {
//初始化
	if Init(){
		for update() {}
	}
	Exit()
}
/*加载配置文件*/
func LoadConfig(fileName string) error {
	if len(fileName) == 0 {
		return fmt.Errorf("配置文件名非法")
	}

	var f *os.File
	var err error
	var canCreateDefaultFile bool=false

	if exchange.IsExist(fileName) {
		f,err= os.Open(fileName)
	} else {
		if canCreateDefaultFile {
			initExchangeConfigDefault(&ExchangeC)
			err=SaveConfig(fileName)
		}else {
			err=fmt.Errorf("配置文件名非法")
		}

		return err
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
	err=json.NewDecoder(f).Decode(&ExchangeC)
	if err != nil{
		if canCreateDefaultFile {
			initExchangeConfigDefault(&ExchangeC)
		}else {
			return err
		}

	}
	return nil
}

/*保存配置文件*/
func SaveConfig(fileName string) error {
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
	err=encoder.Encode(&ExchangeC)
	//f.Sync()
	return err
}
/*获得交易所实例*/
func getExchangeInstance() (exchange.IExchange,error){

	var tact exchange.ITactics
	switch ExchangeC.Tactics {
	case 1:
		tact=&exchange.Tactics1{}
	case 2:
		tact=&exchange.Tactics2{}
	case 3:
		tact=&exchange.Tactics3{}
	case 4:
		tact=&exchange.Tactics4{}
	case 5:
		tact=&exchange.Tactics5{}
	}

	excha:=exchange.Exchange{ExchangeC.Name,ExchangeC.Fee,ExchangeC.APIKey,
	ExchangeC.APISecretKey,ExchangeC.CurTP, tact}

	var ie exchange.IExchange
	if ExchangeC.IsDebug {
		ie=&exchange.BinanceDebugEx{Exchange:excha}
	}else{
		ie=&exchange.BinanceEx{Exchange:excha}
	}
	//

	return ie,nil
}
func initExchangeConfigDefault(coc *ExchangeConfig){

	}

var logFile *os.File

func openLogFile(){
	if logFile!=nil {
		return
	}
	var err error
	lfName:=ExchangeC.LogDir+logFileName
	logFile,err= os.Create(lfName)

	if err != nil {
		log.Infof("创建日志文件失败:%s",lfName)
		return
	}
	log.Infof("创建日志文件:%s",lfName)
}
func closeLogFile(){
	if logFile==nil {
		return
	}



	lfName:=ExchangeC.LogDir+logFileName
	//
	log.Infof("保存日志文件:%s",lfName)
	logFile=nil
}