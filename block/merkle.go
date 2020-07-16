package block

import (
	"encoding/hex"
	"errors"
	"btcnetwork/common"
)

type MerkleNode struct {
	Value string
	Left *MerkleNode
	Right *MerkleNode
}

//构建merkle root hash
func ConstructMerkleRoot(txids []string) (*MerkleNode, error) {
	//参数校验
	if len(txids) == 0 {
		return nil,errors.New("txids is empth")
	}

	//开始构建
	var leafNodes []MerkleNode
	for _, txid := range txids {
		var oneLeaf = MerkleNode{Value:txid, Left:nil, Right:nil}
		leafNodes = append(leafNodes, oneLeaf)
	}
	return ConstructMerkleTreeNodes(leafNodes)
}

func ConstructMerkleTreeNodes(nodes []MerkleNode) (*MerkleNode, error) {
	//参数校验
	if len(nodes) == 0 {
		return nil,errors.New("number of nodes is 0")
	}
	if len(nodes) == 1 {
		return &nodes[0],nil
	}
	if len(nodes) == 2 {
		var root = MerkleNode{}
		var err error
		if root,err = Merge(&nodes[0], &nodes[1]); err != nil {
			return nil, err
		}
		return &root,nil
	}

	var (
		err error
		nodeAmount = len(nodes)
		parentNodes []MerkleNode
		left *MerkleNode
		right *MerkleNode
	)

	//处理三个及三个以上的节点
	for i:=0; i< nodeAmount; i += 2 {
		var parentNode = MerkleNode{}
		left = &nodes[i]
		if i == nodeAmount-1 {//最后一个单独的节点
			right = &nodes[i]
		} else {
			right = &nodes[i+1]
		}
		if parentNode,err = Merge(left, right); err != nil {
			return nil, err
		}
		parentNodes = append(parentNodes,parentNode)
	}
	return ConstructMerkleTreeNodes(parentNodes)
}

func Merge(left *MerkleNode, right *MerkleNode) (MerkleNode, error) {
	var (
		parentNode = MerkleNode{Value:"", Left:left, Right:right}
		data []byte
		err error
		leftBytes []byte
		rightBytes []byte
	)
	//计算的时候txid从BigEdian转变成LittleEdian
	if leftBytes,err = common.ReverseBigEdianString(left.Value); err != nil {
		return MerkleNode{},err
	}
	if rightBytes,err = common.ReverseBigEdianString(right.Value); err != nil {
		return MerkleNode{},err
	}
	data = append(leftBytes, rightBytes...)
	var parentHash = common.Sha256AfterSha256(data)

	//用字符串显示时保存为小端
	data = common.ReverseBytes(parentHash[:])
	parentNode.Value = hex.EncodeToString(data)
	return parentNode,nil
}
