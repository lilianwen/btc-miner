package block

import (
	"errors"
	"math/big"
)

func Bits2Target(bits uint32) *big.Int {
	//解析bits
	factor := bits &0x00ffffff
	power:= bits>>24

	//计算 target = factor * 2^(8*(power-3))
	power = (power-3)*8
	var target = big.NewInt(1)
	target = target.Lsh(target, uint(power))
	target = target.Mul(target, big.NewInt(int64(factor)))
	if len(target.Bytes()) > 32 {//修复计算出来的target值过大，超出32字节长度的bug
		buf := target.Bytes()
		target.SetBytes(buf[:32])
	}
	return target
}

func BigIntTo256Must(target *big.Int) []byte {
	var (
		targetIn256 []byte
		i = 0
		module = new(big.Int)
		quotient *big.Int
		t = new(big.Int).SetBytes(target.Bytes())
	)

	for t.Cmp(big.NewInt(0)) != 0 {
		quotient, module = t.DivMod(t, big.NewInt(256), module)
		targetIn256 = append(targetIn256, byte(module.Uint64()))
		t = quotient
		i++
	}
	return targetIn256
}
func BigIntTo256(target *big.Int) ([]byte, error) {
	//参数校验
	if target.BitLen() > 256 {
		return nil, errors.New("input bug int length in bit is greater than 256")
	}

	return BigIntTo256Must(target),nil
}
func ZipTarget(target *big.Int) ([]byte,error) {
	/*
		将数字转换为256进制数
		如果第一位数字大于127(0x7f),则前面添加0
		压缩结果中的第一位存放该256进制数的位数
		后面三个数存放该256进制数的前三位，如果不足三位，则后面补零
	*/
	var (
		targetIn256 []byte
		err error
		right3Bytes [3]byte
		leftOneByte byte
		bAddOne  = false
	)
	if targetIn256,err = BigIntTo256(target);err != nil {
		return nil,err
	}
	leftOneByte = byte(len(targetIn256))
	if leftOneByte >= 3 {
		if targetIn256[leftOneByte-1] > byte(127) {
			bAddOne = true
			//right3Bytes[2] = 0x00 //默认就是0，所以可以不赋值
			right3Bytes[1] = targetIn256[leftOneByte-1]
			right3Bytes[0] = targetIn256[leftOneByte-2]
		} else {
			right3Bytes[2] = targetIn256[leftOneByte-1]
			right3Bytes[1] = targetIn256[leftOneByte-2]
			right3Bytes[0] = targetIn256[leftOneByte-3]
		}
	} else {
		var startIndex = 0
		if targetIn256[leftOneByte-1] > byte(127) {
			bAddOne = true
			startIndex = 1
		}

		var j=0
		for i:=startIndex; i>=0; i-- {
			targetIn256[i] = targetIn256[int(leftOneByte)-1-j]
			j++
		}
	}

	if bAddOne {
		leftOneByte++
	}
	return append(right3Bytes[:], leftOneByte),nil
}

