package db

import "log"

const (
	BTREE_NODE_TYPE_INTERNAL = 1
	BTREE_NODE_TYPE_LEAF     = 2
	BTREE_PAGE_SIZE          = 4096 // bytes
	BTREE_MAX_KEY_SIZE       = 1000 // bytes
	BTREE_MAX_VAL_SIZE       = 3000 // bytes
)

func init() {
	// Worst-case (node with 1 key-value pair):
	// 2B type, 2B nkeys, 1x8B pointers, 1x2B offsets, 1*4B key/value sizes, xB key, yB value.
	nodeSize := 2 + 2 + 8 + 2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VAL_SIZE
	if nodeSize > BTREE_PAGE_SIZE {
		log.Panicf("tree config incompatible, worst-case node %d exceeds page size %d", nodeSize, BTREE_PAGE_SIZE)
	}
}

type Node struct {
	keys     [][]byte
	children []*Node  // Internal nodes only
	values   [][]byte // Leaf nodes only
}

type KeyValue struct {
	key, value []byte
}
