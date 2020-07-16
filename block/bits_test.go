package block

import (
	"encoding/hex"
	"math/big"
	"btcnetwork/common"
	"reflect"
	"testing"
)

func TestBits2Target(t *testing.T) {
	//t.SkipNow()
	t.Run("test calculate target from bits", func(t *testing.T) {
		var testcases = []struct {
			bits uint32
			target string
		}{
			{486604799, "00000000ffff0000000000000000000000000000000000000000000000000000"},
			{0x1903a30c, "0000000000000003A30C00000000000000000000000000000000000000000000"},
			{0x17103a15, "000000000000000000103a150000000000000000000000000000000000000000"},
			{0xffff7f20, "ff7f200000000000000000000000000000000000000000000000000000000000"},
		}

		for _, elem := range testcases {
			got  := Bits2Target(elem.bits)
			targetInBytes, _ := hex.DecodeString(elem.target)
			want := new(big.Int).SetBytes(targetInBytes)
			if got.Cmp(want) != 0  {
				t.Error("wrong target")
				t.Error("want target:", hex.EncodeToString(want.Bytes()))
				t.Error("got target: ", hex.EncodeToString(got.Bytes()))
			}
		}
	})
}

func TestZipTarget(t *testing.T) {
	//t.SkipNow()
	t.Run("test zip target", func(t *testing.T) {
		var testcases = []struct {
			target string
			want []byte
		}{
			{"03e8", common.Uint32ToBytes(0x0203e800)},
			{"00000000ffff0000000000000000000000000000000000000000000000000000", common.Uint32ToBytes(0x1d00ffff)},
			{"0000000000000003A30C00000000000000000000000000000000000000000000", common.Uint32ToBytes(0x1903a30c)},
			{"ff7f200000000000000000000000000000000000000000000000000000000000", common.Uint32ToBytes(0xffff7f20)},
		}

		var (
			targetArray []byte
			err error
			got []byte
		)
		for _,elem := range testcases {
			if targetArray, err = hex.DecodeString(elem.target);err != nil {//坑爹的，字符串长度一定要是双数
				t.Error(err)
				return
			}
			if got, err = ZipTarget(new(big.Int).SetBytes(targetArray)); err != nil {
				t.Error(err)
				return
			}
			if reflect.DeepEqual(got, elem.want) {
				t.Error("wrong bits")
				t.Error("want: ", hex.EncodeToString(elem.want))
				t.Error("got : ", hex.EncodeToString(got))
			}
		}
	})
}

