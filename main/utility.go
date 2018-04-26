package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/forchain/TradeBot/exchange"
	"github.com/forchain/TradeBot/helpers"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var logFileName string = "null"

/*交易市场的配置*/
type ExchangeConfig struct {
	Name         string
	Fee          float64
	APIKey       string
	APISecretKey string
	Tactics      int
	LogDir       string
	IsDebug      bool
	StopLoss     float64
	StopGain     float64
	CurTP        exchange.TradePairConfig
	MailCfg      helpers.MailInfo
}

type LogFormatter struct {
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {

	temp := fmt.Sprintf("%s:%s\n", time.Now().Format("2006-01-02 15:04:05"),
		entry.Message)
	if logFile != nil {
		logFile.WriteString(temp)
	}
	return []byte(temp), nil
}

var ExchangeC ExchangeConfig
var curExchange exchange.IExchange

func Init() bool {
	//flag
	configFileName := flag.String("cfgFile", "data/config.json", "配置文件路径")
	debugStartTime := flag.String("debugSTime", "", "")
	debugEndTime := flag.String("debugETime", "", "")
	flag.Parse()

	//
	log.SetFormatter(new(LogFormatter))
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	//
	log.Infoln("欢迎使用自动化交易系统^_^")
	//
	log.Infof("配置文件读取路径：%s", *configFileName)
	//
	err := LoadConfig(*configFileName)
	if err != nil {
		log.Errorln(err)
		return false
	}
	//
	if ExchangeC.IsDebug && (*debugStartTime) != "" {
		theTime, err := time.ParseInLocation("2006-01-02-15:04:05", *debugStartTime, time.Local)
		if err == nil {

			exchange.DebugStartTime = &theTime

			log.Infof("设置开始调试时间成功：%v", theTime)
		} else {
			log.Infof("设置开始调试时间失败，将从最开始记录调试")
		}
		theTime2, err2 := time.ParseInLocation("2006-01-02-15:04:05", *debugEndTime, time.Local)
		if err2 == nil {

			exchange.DebugEndTime = &theTime2

			log.Infof("设置结束调试时间成功：%v", theTime2)
		} else {
			log.Infof("设置结束调试时间失败，将调试所有记录")
		}
	}
	helpers.MInfo = &ExchangeC.MailCfg
	//配置交易策略模块

	curExchange, err = getExchangeInstance()
	openLogFile()

	if err != nil {
		log.Errorln(err)
		return false
	}
	//
	log.Infof("当前市场：%s - 交易对：%s  交易费用：%f", curExchange.GetExchange().Name,
		curExchange.GetExchange().CurTP.Name, curExchange.GetExchange().Fee)

	curExchange.Init()
	//

	return true
	//
}
func Exit() {

	curExchange.Exit()

	//err:=SaveConfig(fileName)

	log.Infoln("退出自动化交易系统^_^")

	closeLogFile()

	/*if err!=nil {
		fmt.Println(err)
	}*/
}
func update() bool {

	//time.Sleep(time.Second)
	err := curExchange.Update()
	if err != nil {
		log.Errorln(err)
		if !ExchangeC.IsDebug {
			helpers.DispatchAbnormalExitNotice(logFileName)
		}

		return false
	}
	//
	return true
}
func Run() {
	//初始化
	if Init() {
		for update() {
		}
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
	var canCreateDefaultFile bool = false

	if exchange.IsExist(fileName) {
		f, err = os.Open(fileName)
	} else {
		if canCreateDefaultFile {
			initExchangeConfigDefault(&ExchangeC)
			err = SaveConfig(fileName)
		} else {
			err = fmt.Errorf("配置文件名非法")
		}

		return err
	}

	if err != nil {
		return err
	}

	defer func() {
		cErr := f.Close()
		if err == nil {
			err = cErr
		}
	}()
	//
	err = json.NewDecoder(f).Decode(&ExchangeC)
	if err != nil {
		if canCreateDefaultFile {
			initExchangeConfigDefault(&ExchangeC)
		} else {
			return err
		}

	}
	return nil
}

/*保存配置文件*/
func SaveConfig(fileName string) error {
	var f *os.File
	var err error
	f, err = os.Create(fileName)

	if err != nil {
		return err
	}

	defer func() {
		cErr := f.Close()
		if err == nil {
			err = cErr
		}
	}()
	//
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(&ExchangeC)
	//f.Sync()
	return err
}

/*获得交易所实例*/
func getExchangeInstance() (exchange.IExchange, error) {

	var tact exchange.ITactics = exchange.TacticsMap[ExchangeC.Tactics]

	/*switch ExchangeC.Tactics {
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
	case 6:
		tact=&exchange.Tactics6{}
	case 7:
		tact=&exchange.Tactics7{}
	case 8:
		tact=&exchange.Tactics8{}
	}*/
	if tact == nil {
		return nil, fmt.Errorf("没有这个策略:%d", ExchangeC.Tactics)
	}
	excha := exchange.Exchange{ExchangeC.Name, ExchangeC.Fee, ExchangeC.APIKey,
		ExchangeC.APISecretKey, ExchangeC.LogDir, time.Now(),
		ExchangeC.CurTP, tact, ExchangeC.StopLoss, ExchangeC.StopGain}

	var ie exchange.IExchange
	if ExchangeC.IsDebug {
		ie = &exchange.BinanceDebugEx{Exchange: excha}
	} else {
		ie = &exchange.BinanceEx{Exchange: excha}
	}
	//

	return ie, nil
}
func initExchangeConfigDefault(coc *ExchangeConfig) {

}

var logFile *os.File

func openLogFile() {
	if logFile != nil {
		return
	}
	var err error

	logFileName = ExchangeC.LogDir + curExchange.GetExchange().CurTP.Name + "_" + exchange.GetOptFreTimeStr(curExchange.GetExchange().CurTP.OptFrequency) + "_" + curExchange.GetExchange().StartTime.Format("2006-01-02_15-04-05") + ".log"
	logFile, err = os.Create(logFileName)

	if err != nil {
		log.Infof("创建日志文件失败:%s", logFileName)
		return
	}
	log.Infof("创建日志文件:%s", logFileName)
}
func closeLogFile() {
	if logFile == nil {
		return
	}

	//
	log.Infof("保存日志文件:%s", logFileName)
	logFile = nil
}
