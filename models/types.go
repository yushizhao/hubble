package models

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/yushizhao/hubble/config"
)

type User struct {
	Secret string
	Role   string
}

type STATUS struct {
	Time      string
	Connected bool
}

type TRADE struct {
	Exchange   string
	Symbol     string
	Time       string
	TimeArrive string
	Direction  string
	LastPx     float64
	Qty        float64
}

type DEPTH struct {
	Msg_seq    int
	Exchange   string
	Symbol     string
	Time       string
	TimeArrive string
	AskDepth   orderedmap.OrderedMap
	BidDepth   orderedmap.OrderedMap
}

type OutDEPTH struct {
	Msg_seq    int
	Exchange   string
	Symbol     string
	Time       string
	TimeArrive string
	AskDepth   [][2]float64
	BidDepth   [][2]float64
}

type FairValue struct {
	ISSUE_TIME  string
	BTC_RATE    map[string][2]float64
	TOKEN_PRICE map[string][2]float64
}

type Portfolio struct {
	ClientCode     string
	Value          float64
	ValueComponent map[string]float64
	PnL            float64
	PnLComponent   map[string]float64
	Reserve        map[string][3]float64
}

type Account struct {
	Exchange        string
	PhysicalAccount string
	LogicalAccount  []Portfolio
}

type InAccount struct {
	Counter  string
	Exchange string
	Asset    map[string]map[string][2]float64
}

type GalaxyStatus struct {
	Active     int
	UpdateTime string
}

type StrategyStatus struct {
	Active     int
	UpdateTime string
}

func (this *DEPTH) Output() (o OutDEPTH, err error) {
	outDEPTH := OutDEPTH{
		Msg_seq:    this.Msg_seq,
		Exchange:   this.Exchange,
		Symbol:     this.Symbol,
		Time:       this.Time,
		TimeArrive: this.TimeArrive,
	}

	var tmpPV [2]float64

	for _, p := range this.AskDepth.Keys() {
		tmpPV[0], err = strconv.ParseFloat(p, 64)
		if err != nil {
			return o, err
		}

		rawV, ok := this.AskDepth.Get(p)
		if !ok {
			return o, fmt.Errorf("this.AskDepth.Get(p) failed at p = %s", p)
		}
		v, ok := rawV.(float64)
		if !ok {
			return o, fmt.Errorf("rawV.(float64) failed at rawV of %T, %v", rawV, rawV)
		}
		tmpPV[1] = v

		outDEPTH.AskDepth = append(outDEPTH.AskDepth, tmpPV)
	}

	for _, p := range this.BidDepth.Keys() {
		tmpPV[0], err = strconv.ParseFloat(p, 64)
		if err != nil {
			return o, err
		}

		rawV, ok := this.BidDepth.Get(p)
		if !ok {
			return o, fmt.Errorf("this.BidDepth.Get(p) failed at p = %s", p)
		}
		v, ok := rawV.(float64)
		if !ok {
			return o, fmt.Errorf("rawV.(float64) failed at rawV of %T, %v", rawV, rawV)
		}
		tmpPV[1] = v

		outDEPTH.BidDepth = append(outDEPTH.BidDepth, tmpPV)
	}

	return outDEPTH, err
}

func (this *InAccount) ToAccount() Account {
	var portfolios []Portfolio
	for asset, portfoliosMapBalance := range this.Asset {
		for portfolioName, balance := range portfoliosMapBalance {
			hasFound := false
			for _, p := range portfolios {
				if p.ClientCode == portfolioName {
					hasFound = true
					var tmp [3]float64
					tmp[0] = balance[0]
					tmp[1] = balance[1]
					tmp[2] = balance[0] + balance[1]
					// made in !hasFound
					p.Reserve[asset] = tmp
				}
			}
			if !hasFound {
				p := Portfolio{portfolioName, 0.0, make(map[string]float64), 0.0, make(map[string]float64), make(map[string][3]float64)}
				var tmp [3]float64
				tmp[0] = balance[0]
				tmp[1] = balance[1]
				tmp[2] = balance[0] + balance[1]
				// made in !hasFound
				p.Reserve[asset] = tmp
				// may cache a list of p to be appended after this asset
				// because one p name will not appear twice under one asset
				portfolios = append(portfolios, p)
			}
		}
	}
	return Account{this.Exchange, this.Counter, portfolios}
}

// init PnL as well
func (this *Account) EstimateValue(fairValue FairValue) error {

	for i, _ := range this.LogicalAccount {

		value := 0.0

		for k, v := range this.LogicalAccount[i].Reserve {

			if k == "BTC" {
				value = value + v[2]
				// made in ToAccount
				this.LogicalAccount[i].ValueComponent[k] = v[2]
				this.LogicalAccount[i].PnLComponent[k] = v[2]
				continue
			}

			if pv, ok := fairValue.BTC_RATE[k]; ok {
				tmp := v[2] / pv[0]
				value = value + tmp
				// made in ToAccount
				this.LogicalAccount[i].ValueComponent[k] = tmp
				this.LogicalAccount[i].PnLComponent[k] = tmp
				continue
			}

			if pv, ok := fairValue.TOKEN_PRICE[k]; ok {
				tmp := v[2] * pv[0]
				value = value + tmp
				// made in ToAccount
				this.LogicalAccount[i].ValueComponent[k] = tmp
				this.LogicalAccount[i].PnLComponent[k] = tmp
				continue
			}
		}
		this.LogicalAccount[i].Value = value
		this.LogicalAccount[i].PnL = value
	}

	return nil
}

func (this *Portfolio) calculatePnl(that Portfolio) error {
	if this.ClientCode == that.ClientCode {
		this.PnL = this.Value - that.Value

		for k, v := range this.ValueComponent {
			if vthat, ok := that.ValueComponent[k]; ok {
				this.PnLComponent[k] = v - vthat
			} else {
				this.PnLComponent[k] = v
			}
		}

		// if one position closed
		for k, v := range that.ValueComponent {
			if _, ok := this.ValueComponent[k]; !ok {
				this.PnLComponent[k] = -v
			}
		}
	}
	return nil
}

func (this *Account) CalculatePnl(that Account) error {
	if this.Exchange == that.Exchange {
		for i, _ := range this.LogicalAccount {
			for _, pin := range that.LogicalAccount {
				err := this.LogicalAccount[i].calculatePnl(pin)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (this *Portfolio) add(that Portfolio) error {

	this.Value = this.Value + that.Value
	this.PnL = this.PnL + that.PnL

	for k, v := range that.Reserve {
		if _, ok := this.Reserve[k]; ok {
			var tmp [3]float64
			tmp[0] = this.Reserve[k][0] + v[0]
			tmp[1] = this.Reserve[k][1] + v[1]
			tmp[2] = this.Reserve[k][2] + v[2]
			this.Reserve[k] = tmp
		} else {
			this.Reserve[k] = v
		}
	}

	// may combined into above loop, because keys are the same
	for k, v := range that.ValueComponent {
		if _, ok := this.ValueComponent[k]; ok {
			this.ValueComponent[k] = this.ValueComponent[k] + v
		} else {
			this.ValueComponent[k] = v
		}
	}

	for k, v := range that.PnLComponent {
		if _, ok := this.PnLComponent[k]; ok {
			this.PnLComponent[k] = this.PnLComponent[k] + v
		} else {
			this.PnLComponent[k] = v
		}
	}

	return nil
}

func (this *Portfolio) copy() (that Portfolio, err error) {
	var mod bytes.Buffer
	enc := gob.NewEncoder(&mod)
	dec := gob.NewDecoder(&mod)
	err = enc.Encode(this)
	if err != nil {
		return that, err
	}
	err = dec.Decode(&that)
	// for formatting
	if err != nil {
		return that, err
	}
	return that, err
}

func (this *Account) LogicalTotal() error {
	total := Portfolio{"TOTAL", 0.0, make(map[string]float64), 0.0, make(map[string]float64), make(map[string][3]float64)}

	for _, p := range this.LogicalAccount {
		if p.ClientCode == "TOTAL" {
			return fmt.Errorf("already has TOTAL")
		}

		err := total.add(p)
		if err != nil {
			return err
		}
	}

	this.LogicalAccount = append(this.LogicalAccount, total)
	return nil
}

func PhysicalTotal(these []Account) ([]Account, error) {
	var portfolio []Portfolio
	for _, a := range these {

		if a.PhysicalAccount == "TOTAL" {
			return these, fmt.Errorf("already has TOTAL")
		}
		for _, pin := range a.LogicalAccount {

			hasFound := false
			for i, _ := range portfolio {
				if pin.ClientCode == portfolio[i].ClientCode {
					hasFound = true
					err := portfolio[i].add(pin)
					if err != nil {
						return nil, err
					}
				}
			}

			if !hasFound {
				pinCopy, err := pin.copy() // a deep copy
				if err != nil {
					return nil, err
				}
				portfolio = append(portfolio, pinCopy)
			}

		}

	}
	total := Account{"TOTAL", "TOTAL", portfolio}
	these = append(these, total)
	return these, nil
}

type StrategySummary struct {
	StrategyName string
	Active       int
	Position     map[string]float64
	Account      map[string]float64
	UpdateTime   string
}

type StrategyPosition struct {
	StrategyName string
	InstrumentID string
	Position     float64
	UpdateTime   string
}

type StrategyAccount struct {
	StrategyName string
	InstrumentID string
	Account      float64
	UpdateTime   string
}

type StrategyMarket struct {
	StrategyName string
	InstrumentID string
	AskPrice1    float64
	BidPrice1    float64
	AskVolume1   float64
	BidVolume1   float64
	UpdateTime   string
}

type StrategyUserDefineInput struct {
	StrategyName string
	Key          string
	Value        string
	UpdateTime   string
}

type StrategyUserDefine struct {
	StrategyName string
	UserDefine   map[string]string
	UpdateTime   string
}

type StrategyTrade struct {
	StrategyName    string
	InstrumentID    string
	Direction       string
	TradePrice      float64
	TradeVolume     float64
	BaseCurrency    string
	Fee             float64
	StrategyOrderID string
	UpdateTime      string
}

type StrategyOrder struct {
	StrategyName    string
	InstrumentID    string
	Direction       string
	Price           float64
	Volume          float64
	TradePrice      float64
	TradeVolume     float64
	StrategyOrderID string
	OrderStatus     string
	UpdateTime      string
}

type StrategyMessageSet struct {
	UpdateTimestamp int64
	Summary         StrategySummary
	Market          map[string]StrategyMarket
	UserDefine      []StrategyUserDefine
	Trade           []StrategyTrade
	Order           []StrategyOrder
}

func MakeStrategyMessageSet() *StrategyMessageSet {
	set := new(StrategyMessageSet)

	summary := new(StrategySummary)
	summary.Position = make(map[string]float64)
	summary.Account = make(map[string]float64)
	set.Summary = *summary

	set.Market = make(map[string]StrategyMarket)

	return set
}

func (set *StrategyMessageSet) InsertSummary(that StrategySummary) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	set.Summary.Active = that.Active

	return nil
}

func (set *StrategyMessageSet) InsertPosition(that StrategyPosition) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	set.Summary.Position[that.InstrumentID] = that.Position

	return nil
}

func (set *StrategyMessageSet) InsertAccount(that StrategyAccount) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	set.Summary.Account[that.InstrumentID] = that.Account

	return nil
}

func (set *StrategyMessageSet) InsertMarket(that StrategyMarket) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	set.Market[that.InstrumentID] = that

	return nil
}

func (this StrategyUserDefineInput) Transform() (that StrategyUserDefine) {
	that.StrategyName = this.StrategyName
	that.UpdateTime = this.UpdateTime
	that.UserDefine = make(map[string]string)
	that.UserDefine[this.Key] = this.Value
	return that
}

func (set *StrategyMessageSet) InsertUserDefine(that StrategyUserDefine) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	if len(set.UserDefine) == config.Server.UserDefineLimit {
		set.UserDefine = append([]StrategyUserDefine{that}, set.UserDefine[:(config.Server.UserDefineLimit-1)]...)
	} else {
		set.UserDefine = append([]StrategyUserDefine{that}, set.UserDefine...)
	}

	return nil
}

func (set *StrategyMessageSet) InsertTrade(that StrategyTrade) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	if len(set.Trade) == config.Server.TradeLimit {
		set.Trade = append([]StrategyTrade{that}, set.Trade[:(config.Server.TradeLimit-1)]...)
	} else {
		set.Trade = append([]StrategyTrade{that}, set.Trade...)
	}

	return nil
}

func (set *StrategyMessageSet) InsertOrder(that StrategyOrder) error {
	t, err := time.Parse(updateTimeLayout, that.UpdateTime)
	if err != nil {
		return err
	}
	set.UpdateTimestamp = t.Unix()

	NewOrder := []StrategyOrder{that}

	hasFound := false
	for i, v := range set.Order {
		if v.StrategyOrderID == that.StrategyOrderID {
			NewOrder = append(NewOrder, set.Order[:i]...)
			NewOrder = append(NewOrder, set.Order[(i+1):]...)
			hasFound = true
			break
		}
	}

	if hasFound {
		set.Order = NewOrder
	} else {
		if len(set.Order) == config.Server.OrderLimit {
			set.Order = append(NewOrder, set.Order[:(config.Server.TradeLimit-1)]...)
		} else {
			set.Order = append(NewOrder, set.Order...)
		}
	}

	return nil
}
