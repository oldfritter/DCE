package models

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

type OrderCurrency struct {
	Fee        decimal.Decimal `json:"fee"`
	Currency   string          `json:"currency"`
	CurrencyId int             `json:"currency_id"`
	Fixed      int             `json:"fixed"`
}

type Market struct {
	CommonModel
	Name            string          `json:"name"" gorm:"type:varchar(16)"`
	Code            string          `json:"code" gorm:"type:varchar(16)"`
	PriceGroupFixed int             `json:"price_group_fixed"`
	SortOrder       int             `json:"sort_order"`
	AskCurrencyId   int             `json:"ask_currency_id"`
	BidCurrencyId   int             `json:"bid_currency_id"`
	AskFee          decimal.Decimal `json:"ask_fee" gorm:"type:decimal(32,16);default:null;"`
	BidFee          decimal.Decimal `json:"bid_fee" gorm:"type:decimal(32,16);default:null;"`
	AskFixed        int             `json:"ask_fixed"`
	BidFixed        int             `json:"bid_fixed"`
	Visible         bool            `json:"visible"`
	Tradable        bool            `json:"tradable"`

	// 暂存数据
	Ticker       TickerAspect  `sql:"-" json:"ticker"`
	LatestKLines map[int]KLine `sql:"-" json:"-"`

	// 撮合相关属性
	Ack             bool   `json:"-"`
	Durable         bool   `json:"-"`
	MatchingAble    bool   `json:"-"`
	MatchingNode    string `json:"-" gorm:"default:'a'; type:varchar(11)"`
	TradeTreatNode  string `json:"-" gorm:"default:'a'; type:varchar(11)"`
	OrderCancelNode string `json:"-" gorm:"default:'a'; type:varchar(11)"`
	Running         bool   `json:"-" sql:"-"`

	Matching    string `json:"-"`
	TradeTreat  string `json:"-"`
	OrderCancel string `json:"-"`
}

var AllMarkets []Market

func InitAllMarkets(db *utils.GormDB) {
	db.Where("visible = ?", true).Find(&AllMarkets)
}

func FindAllMarket() []Market {
	return AllMarkets
}

func FindMarketById(id int) (Market, error) {
	for _, market := range AllMarkets {
		if market.Id == id {
			return market, nil
		}
	}
	return Market{}, fmt.Errorf("No market can be found.")
}

func FindMarketByCode(code string) (Market, error) {
	for _, market := range AllMarkets {
		if market.Code == code {
			return market, nil
		}
	}
	return Market{}, fmt.Errorf("No market can be found.")
}

func (market *Market) AfterCreate(db *gorm.DB) {
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	ticker := Ticker{MarketId: market.Id, Name: market.Name}
	b, _ := json.Marshal(ticker)
	if _, err := tickerRedis.Do("HSET", TickersRedisKey, market.Id, string(b)); err != nil {
		fmt.Println("{ error: ", err, "}")
		return
	}
}

func (market *Market) AfterFind(db *gorm.DB) {
	market.LatestKLines = make(map[int]KLine)
}

// Exchange
func (assignment *Market) MatchingExchange() string {
	return assignment.Matching
}
func (assignment *Market) TradeTreatExchange() string {
	return assignment.TradeTreat
}
func (assignment *Market) OrderCancelExchange() string {
	return assignment.OrderCancel
}

// Queue
func (assignment *Market) MatchingQueue() string {
	return assignment.MatchingExchange() + "." + assignment.Code
}
func (assignment *Market) TradeTreatQueue() string {
	return assignment.TradeTreatExchange() + "." + assignment.Code
}
func (assignment *Market) OrderCancelQueue() string {
	return assignment.OrderCancelExchange() + "." + assignment.Code
}

func (market *Market) LatestTradesRedisKey() string {
	return fmt.Sprintf("goDCE:latestTrades:%v", market.Code)
}
func (market *Market) TickerRedisKey() string {
	return "goDCE:ticker:" + market.Code
}
func (market *Market) KLineRedisKey(period int64) string {
	return fmt.Sprintf("goDCE:k:%v:%v", market.Id, period)
}

func (market *Market) AskRedisKey() string {
	return fmt.Sprintf("goDCE:depth:%v:ask", market.Id)
}
func (market *Market) BidRedisKey() string {
	return fmt.Sprintf("goDCE:depth:%v:bid", market.Id)
}

// Notify
func (market *Market) KLineNotify(period int64) string {
	return "market:kLine:notify"
}
func (market *Market) TickerNotify() string {
	return fmt.Sprintf("market:ticker:notify:%v", market.Id)
}
