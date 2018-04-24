package exchange

import (
	"time"
	"github.com/Akagi201/utilgo/enums"
)

type TacticsData struct {
	CurRecords *ExchangeRecords
	CurMACDs *MACDDatas
	Excha IExchange
	account *Account
}

type ITactics interface {
	Init(date *TacticsData)
	Update()
	GetID() int
}

const (

)
type OptRecord struct {
	optType int
	reason int
	time time.Time
	orderID int64
	indexOpt int64
	price float64
}

var (
	OrderStatus   enums.Enum
	OrderNew  = OrderStatus.Iota("NEW")
	OrderPartiallyFilled=OrderStatus.Iota("PARTIALLY_FILLED")
	OrderFilled = OrderStatus.Iota("FILLED")
	OrderCanceled=OrderStatus.Iota("CANCELED")
	OrderRejected=OrderStatus.Iota("REJECTED")
	OrderExpired=OrderStatus.Iota("EXPIRED")
)

var (
	TradeType   enums.Enum
	TradeBuy= TradeType.Iota("buy")
	TradeSell = TradeType.Iota("sell")
)

const (
	DIF_UP_0=iota
	DIF_DOWN_0
	FALLAWAY_TOP
	FALLAWAY_BOTTOM
	GOLDEN_CROSS
	DEAD_CROSS
	STOP_LOSS
	STOP_GAIN
)

var reasonStr=[]string{"DIF线上穿MACD 0轴","DIF线下穿MACD 0轴","顶背离","底背离","黄金金叉","黄金死叉","当前价已经低于止损价","接回仓位"}