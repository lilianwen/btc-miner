package p2p

import (
	"btcnetwork/common"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
	"log"
)

type GetblocksPayload struct {
	Version uint32
	HashCount common.VarInt
	HashStart [32]byte
	HashStop [32]byte
}

func (gp *GetblocksPayload)Serialize() []byte {
	var buf []byte
	var uint32Buf [4]byte

	binary.LittleEndian.PutUint32(uint32Buf[:], gp.Version)

	buf = append(buf, uint32Buf[:]...)
	buf = append(buf, gp.HashCount.Data...)
	buf = append(buf, gp.HashStart[:]...)
	buf = append(buf, gp.HashStop[:]...)

	return buf
}

func (gp *GetblocksPayload)Parse(data []byte) error {
	if len(data) < gp.Len() {
		return errors.New("data is too short for getblocks payload")
	}
	binary.LittleEndian.Uint32(data[:4])
	if err := gp.HashCount.Parse(hex.EncodeToString(data[4:])); err != nil {
		return err
	}
	start := 4 + gp.HashCount.Len()
	copy(gp.HashStart[:], data[start:start+32])
	start += 32
	copy(gp.HashStop[:], data[start:start+32])
	return nil
}

func (gp *GetblocksPayload)Len() int {
	return 68+gp.HashCount.Len()
}

func (node *Node)HandleGetblocks(peer *Peer, payload []byte) error {
	gp := GetblocksPayload{}
	err := gp.Parse(payload)
	if err != nil {
		return err
	}
	log.Println("get hash count:", gp.HashCount.Value)
	log.Println("hash start:", hex.EncodeToString(gp.HashStart[:]))
	log.Println("hash stop:", hex.EncodeToString(gp.HashStop[:]))

	//对方节点告诉我他的最新的区块hash值是多少，他想要的最新的区块哈希值是多少（如果是全0）表示他想要我的最新的区块
	//我对比一下我这边的最新区块哈希值和它的是不是相同，如果相同表示我们拥有相同的区块数据，
	//如果不同，则我的更旧，我就向它要区块（发送getblocks消息），如果我的更新，我就发送inv消息给他，向它提供最新的区块
	latestBlockHash := node.GetLatestBlockHash()
	hashStart := hex.EncodeToString(gp.HashStart[:])
	if latestBlockHash != hashStart {
		if node.HasBlockHash(hashStart) { //说明我的区块较对方节点更新
			// todo:把我的区块发送给他
			return errors.New("not support sending blocks to remote peers")
		} else {
			// todo:向对方请求最新的区块
			return errors.New("not support receiving blocks from remote peers")
		}

	}

	//发送inv消息表明我拥有的区块和对方一样
	invp := InvPayload{}
	invp.Count = common.NewVarInt(0)
	msg, err := NewMsg("inv", invp.Serialize())
	if err != nil {
		return err
	}
	if err = node.sendMsg(peer.Conn, msg.Serialize()); err != nil {
		return err
	}

	return nil

}

// todo: 获取本地最新的区块哈希
func (node *Node)GetLatestBlockHash() string {
	return "f67ad7695d9b662a72ff3d8edbbb2de0bfa67b13974bb9910d116d5cbd863e68"
}

// todo:判断本地是否有该哈希值
func (node *Node)HasBlockHash(hash string) bool {
	return false
}