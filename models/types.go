package models

import (
	"fmt"
	"strconv"

	"github.com/iancoleman/orderedmap"
)

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
	ClientCode string
	Value      float64
	PnL        float64
	Reserve    map[string][3]float64
}

type Account struct {
	PhysicalAccount string
	LogicalAccount  []Portfolio
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
