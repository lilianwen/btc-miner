package block

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
)

type BlockDiff struct {
	Height float64 `json:"height"`
	Bits float64 `json:"bits"`
	Timestamp float64 `json:"timestamp"`
}

func CalNextTarget(start, end BlockDiff) (*big.Int,error) {
	if int64(start.Height) % 2016 != 0 {
		return nil,errors.New("start height % 2016 should be 0")
	}
	if end.Height != (start.Height+2015) {
		return nil,errors.New("end height should == start height + 2016")
	}

	timeVal := float64( end.Timestamp - start.Timestamp ) //单位：秒
	fmt.Println("actual timeval:", timeVal)
	expectTimeVal := float64( 2016*10*60 ) //理想情况下2016个区块之间时间间隔，单位：秒
	fmt.Println("expect timeval:", expectTimeVal)
	rate := float64(timeVal)/float64(expectTimeVal)
	if rate > 4 {
		rate = 4
	}
	if rate < 0.25 {
		rate = 0.25
	}
	var (
		currTarget *big.Int
		nextTarget *big.Int
		maxTargetInBigInt *big.Int
		maxTarget decimal.Decimal
	)
	currTarget = Bits2Target(uint32(end.Bits))

	maxTargetInBigInt,_ = new(big.Int).SetString("00000000ffff0000000000000000000000000000000000000000000000000000", 16)
	maxTarget = decimal.NewFromBigInt(maxTargetInBigInt, 0)
	ret := decimal.NewFromBigInt(currTarget, 0).Mul(decimal.NewFromFloat(rate))//
	if ret.GreaterThan(maxTarget) {//精度都不一样导致比较出错
		nextTarget = maxTarget.BigInt()
	} else {
		nextTarget = ret.BigInt()
	}
	return nextTarget, nil
}
