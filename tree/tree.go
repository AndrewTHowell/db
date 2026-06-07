package tree

import (
	"log"
)

const (
	NODE_TYPE_INTERNAL = 1
	NODE_TYPE_LEAF     = 2
	PAGE_SIZE          = 4096 // bytes
	MAX_KEY_SIZE       = 1000 // bytes
	MAX_VAL_SIZE       = 3000 // bytes
)

func init() {
	// Worst-case (node with 1 key-value pair):
	// 2B type, 2B nkeys, 1x8B pointers, 1x2B offsets, 1*4B key/value sizes, xB key, yB value.
	nodeSize := 2 + 2 + 8 + 2 + 4 + MAX_KEY_SIZE + MAX_VAL_SIZE
	if nodeSize > PAGE_SIZE {
		log.Panicf("tree config incompatible, worst-case node %d exceeds page size %d", nodeSize, PAGE_SIZE)
	}
}

func LeafInsert(newNode, oldNode BNode, idx uint16, key, value []byte) {
	newNode.SetHeader(NODE_TYPE_LEAF, oldNode.nkeys()+1)
	NodeAppendRange(newNode, oldNode, 0, 0, idx)
	NodeAppendKeyValue(newNode, idx, 0, key, value)
	NodeAppendRange(newNode, oldNode, idx+1, idx, oldNode.nkeys()-idx)
}

func NodeAppendRange(newNode, oldNode BNode, dstNew, srcOld, n uint16) {
	for i := range n {
		dst, src := dstNew+i, srcOld+i
		NodeAppendKeyValue(newNode, dst, oldNode.getPtr(src), oldNode.GetKey(src), oldNode.GetValue(src))
	}
}

func NodeAppendKeyValue(node BNode, idx uint16, ptr uint64, key, value []byte) {
	node.setPtr(idx, ptr)
	keyValue := newKeyValue(node[node.keyValuePosition(idx):], key, value)
	// Set offset for next key.
	node.setOffset(idx+1, node.getOffset(idx)+uint16(len(keyValue)))
}
