package common

import (
	"errors"
	"math/big"
)

func Base58Encode(data []byte) string {
	var (
		alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
		bigNum   = new(big.Int).SetBytes(data)
		divisor  = big.NewInt(58)
		modulus  = big.NewInt(0)
		result   string
	)
	for {
		if bigNum.Cmp(big.NewInt(0)) == 0 {
			break
		}
		bigNum.DivMod(bigNum, divisor, modulus)
		result = string(alphabet[modulus.Int64()]) + result
	}
	return result
}

func Base58Decode(data string) ([]byte, error) {
	var alphabetMap = map[string]int64{
		"1": 0,
		"2": 1,
		"3": 2,
		"4": 3,
		"5": 4,
		"6": 5,
		"7": 6,
		"8": 7,
		"9": 8,
		"A": 9,
		"B": 10,
		"C": 11,
		"D": 12,
		"E": 13,
		"F": 14,
		"G": 15,
		"H": 16,
		"J": 17,
		"K": 18,
		"L": 19,
		"M": 20,
		"N": 21,
		"P": 22,
		"Q": 23,
		"R": 24,
		"S": 25,
		"T": 26,
		"U": 27,
		"V": 28,
		"W": 29,
		"X": 30,
		"Y": 31,
		"Z": 32,
		"a": 33,
		"b": 34,
		"c": 35,
		"d": 36,
		"e": 37,
		"f": 38,
		"g": 39,
		"h": 40,
		"i": 41,
		"j": 42,
		"k": 43,
		"m": 44,
		"n": 45,
		"o": 46,
		"p": 47,
		"q": 48,
		"r": 49,
		"s": 50,
		"t": 51,
		"u": 52,
		"v": 53,
		"w": 54,
		"x": 55,
		"y": 56,
		"z": 57,
	}
	var (
		bigNum = big.NewInt(0)
	)
	for _, d := range data {
		e, ok := alphabetMap[string(d)]
		if !ok {
			return nil, errors.New("contain invalid aplha")
		}
		//把58进制转换成十进制
		bigNum.Mul(bigNum, big.NewInt(58))
		bigNum.Add(bigNum, new(big.Int).SetInt64(e))
	}

	return bigNum.Bytes(), nil
}
