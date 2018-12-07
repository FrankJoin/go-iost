package host

import (
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
)

// Teller handler of iost
type Teller struct {
	h         *Host
	cost      map[string]contract.Cost
	cacheCost contract.Cost
}

// NewTeller new teller
func NewTeller(h *Host) Teller {
	return Teller{
		h:    h,
		cost: make(map[string]contract.Cost),
	}
}

// Costs ...
func (h *Teller) Costs() map[string]contract.Cost {
	return h.cost
}

// ClearCosts ...
func (h *Teller) ClearCosts() {
	h.cost = make(map[string]contract.Cost)
}

// ClearRAMCosts ...
func (h *Teller) ClearRAMCosts() {
	newCost := make(map[string]contract.Cost)
	for k, c := range h.cost {
		if c.Net != 0 || c.CPU != 0 {
			newCost[k] = contract.NewCost(0, c.Net, c.CPU)
		}
	}
	h.cost = newCost
}

// AddCacheCost ...
func (h *Teller) AddCacheCost(c contract.Cost) {
	h.cacheCost.AddAssign(c)
}

// CacheCost ...
func (h *Teller) CacheCost() contract.Cost {
	return h.cacheCost
}

// FlushCacheCost ...
func (h *Teller) FlushCacheCost() {
	h.PayCost(h.cacheCost, "")
	h.cacheCost = contract.Cost0()
}

// ClearCacheCost ...
func (h *Teller) ClearCacheCost() {
	h.cacheCost = contract.Cost0()
}

// PayCost ...
func (h *Teller) PayCost(c contract.Cost, who string) {
	costMap := make(map[string]contract.Cost)
	if c.CPU > 0 || c.Net > 0 {
		costMap[who] = contract.Cost{CPU: c.CPU, Net: c.Net}
	}
	for _, item := range c.DataList {
		if oc, ok := costMap[item.Payer]; ok {
			oc.AddAssign(contract.Cost{Data: item.Val, DataList: []contract.DataItem{item}})
			costMap[item.Payer] = oc
		} else {
			costMap[item.Payer] = contract.Cost{Data: item.Val, DataList: []contract.DataItem{item}}
		}
	}
	for who, c := range costMap {
		if oc, ok := h.cost[who]; ok {
			oc.AddAssign(c)
			h.cost[who] = oc
		} else {
			h.cost[who] = c
		}
	}
}

// DoPay ...
func (h *Teller) DoPay(witness string, gasRatio int64) error {
	for k, c := range h.cost {
		fee := gasRatio * c.ToGas()
		if fee != 0 {
			gas := &common.Fixed{
				Value:   fee,
				Decimal: 2,
			}
			err := h.h.CostGas(k, gas)
			if err != nil {
				return fmt.Errorf("pay cost failed: %v, %v", k, err)
			}
		}
		// contracts in "iost" domain will not pay for ram
		if !strings.HasSuffix(k, ".iost") {
			var payer string
			if h.h.IsContract(k) {
				p, _ := h.h.GlobalMapGet("system.iost", "contract_owner", k)
				var ok bool
				payer, ok = p.(string)
				if !ok {
					return fmt.Errorf("DoPay failed: contract %v has no owner", k)
				}
			} else {
				payer = k
			}

			ram := c.Data
			currentRAM := h.h.db.TokenBalance("ram", payer)
			ilog.Infof("id: %v, actual %v", payer, currentRAM)
			if currentRAM-ram < 0 {
				return fmt.Errorf("pay ram failed. id: %v need %v, actual %v", payer, ram, currentRAM)
			}
			h.h.db.SetTokenBalance("ram", payer, currentRAM-ram)
		}
	}
	return nil
}

// Privilege ...
func (h *Teller) Privilege(id string) int {
	am, ok := h.h.ctx.Value("auth_list").(map[string]int)
	if !ok {
		return 0
	}
	i, ok := am[id]
	if !ok {
		i = 0
	}
	return i
}
