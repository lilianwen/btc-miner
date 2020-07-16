package block

import (
	//"encoding/binary"
	//"fmt"
	//"generateAddress/utils"
	//"math/big"
	//"testing"
)
/*
func TestCalNextTarget(t *testing.T) {
	//t.SkipNow()
	t.Run("test next target", func(t *testing.T) {
		var (
			start *utils.BlockDiff
			end *utils.BlockDiff
			want *utils.BlockDiff
			nextTarget *big.Int
			gotBits uint32
			err error
			nextTargetInBytes []byte
		)

		for i:=14; i<100; i++ {
			startHeight := i*2016
			fmt.Println("startHeight:", startHeight)
			if start,err = utils.GetBlockInfo(uint64(startHeight));err != nil {
				t.Error(err)
				return
			}
			if end,err = utils.GetBlockInfo(uint64(startHeight+2015));err != nil {
				t.Error(err)
				return
			}
			if want,err = utils.GetBlockInfo(uint64(startHeight+2016));err != nil {
				t.Error(err)
				return
			}

			if nextTarget,err = CalNextTarget(start, end);err != nil {
				t.Error(err)
				return
			}

			if nextTargetInBytes, err = ZipTarget(nextTarget); err != nil {
				t.Error(err)
				return
			}
			gotBits = uint32(binary.LittleEndian.Uint32(nextTargetInBytes))
			if gotBits != uint32(want.Bits) {
				t.Error("wrong bits")
				t.Errorf("want:%x\n", uint32(want.Bits))
				t.Errorf("got:%x\n", gotBits)
				return
			}
		}
	})
}
*/
