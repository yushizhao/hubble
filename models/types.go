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

func (this *Portfolio) EstimateValue(fairValue FairValue) error {
	return nil
}

func (this *Portfolio) CalculatePnl(that Portfolio) error {
	return nil
}

func (this *Account) Complete(that Account, fairValue FairValue) error {
	return nil
}

func (this *Account) LogicalTotal() error {
	clientCode := "TOTAL"
	value := 0.0
	pnl := 0.0
	reserve := make(map[string][3]float64)

	for _, p := range this.LogicalAccount {
		if p.ClientCode == clientCode {
			return fmt.Errorf("already has TOTAL")
		}
		value = value + p.Value
		pnl = pnl + p.PnL

		for k, v := range p.Reserve {
			if _, ok := reserve[k]; ok {
				var tmp [3]float64
				tmp[0] = reserve[k][0] + v[0]
				tmp[1] = reserve[k][1] + v[1]
				tmp[2] = reserve[k][2] + v[2]
				reserve[k] = tmp
			} else {
				reserve[k] = v
			}
		}
	}

	total := Portfolio{clientCode, value, pnl, reserve}
	this.LogicalAccount = append(this.LogicalAccount, total)
	return nil
}

func PhysicalTotal(these []Account) ([]Account, error) {
	physical := "TOTAL"
	var portfolio []Portfolio
	for _, a := range these {
		if a.PhysicalAccount == physical {
			return these, fmt.Errorf("already has TOTAL")
		}
		for _, pin := range a.LogicalAccount {
			hasFound := false
			for i, _ := range portfolio {
				if pin.ClientCode == portfolio[i].ClientCode {
					hasFound = true
					portfolio[i].Value = portfolio[i].Value + pin.Value
					portfolio[i].PnL = portfolio[i].PnL + pin.PnL

					for k, v := range pin.Reserve {
						if _, ok := portfolio[i].Reserve[k]; ok {
							var tmp [3]float64
							tmp[0] = portfolio[i].Reserve[k][0] + v[0]
							tmp[1] = portfolio[i].Reserve[k][1] + v[1]
							tmp[2] = portfolio[i].Reserve[k][2] + v[2]
							portfolio[i].Reserve[k] = tmp
						} else {
							portfolio[i].Reserve[k] = v
						}
					}
				}
			}
			if !hasFound {
				portfolio = append(portfolio, pin)
			}
		}
	}
	total := Account{physical, portfolio}
	these = append(these, total)
	return these, nil
}
