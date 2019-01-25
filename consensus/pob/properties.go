package pob

import (
	"github.com/iost-official/go-iost/ilog"
	"strings"
	"sync"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
)

var staticProperty *StaticProperty

// StaticProperty handles the the static property of pob.
type StaticProperty struct {
	account           *account.KeyPair
	NumberOfWitnesses int64
	mu                sync.RWMutex
}

func newStaticProperty(account *account.KeyPair, number int64) *StaticProperty {
	property := &StaticProperty{
		account:           account,
		NumberOfWitnesses: number,
	}
	return property
}

func (property *StaticProperty) isWitness(w string, witnessList []string) bool {
	for _, v := range witnessList {
		if strings.Compare(v, w) == 0 {
			return true
		}
	}
	return false
}

var (
	second2nanosecond int64 = 1000000000
)

func witnessOfNanoSec(nanosec int64, witnessList []string) string {
	return witnessOfSec(nanosec/second2nanosecond, witnessList)
}

func witnessOfSec(sec int64, witnessList []string) string {
	return witnessOfSlot(sec/common.SlotLength, witnessList)
}

func witnessOfSlot(slot int64, witnessList []string) string {
	index := slot % staticProperty.NumberOfWitnesses
	ilog.Infof("witnessList len:%v ,index: %v ,slot: %v, NumOfWitness: %v", len(witnessList), index, slot, staticProperty.NumberOfWitnesses)
	witness := witnessList[index]
	return witness
}

func slotOfSec(sec int64) int64 {
	return sec / common.SlotLength
}

func timeUntilNextSchedule(timeSec int64) int64 {
	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
	return (currentSlot+1)*second2nanosecond*common.SlotLength - timeSec
}

// GetStaticProperty return property. RPC needs it.
func GetStaticProperty() *StaticProperty {
	return staticProperty
}
