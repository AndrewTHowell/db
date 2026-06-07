package tree

import (
	"bytes"
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

func LeafUpdate(newNode, oldNode BNode, idx uint16, key, value []byte) {
	newNode.SetHeader(NODE_TYPE_LEAF, oldNode.nkeys())
	NodeAppendRange(newNode, oldNode, 0, 0, idx)
	NodeAppendKeyValue(newNode, idx, 0, key, value)
	NodeAppendRange(newNode, oldNode, idx+1, idx+1, oldNode.nkeys()-idx-1)
}

func NodeSplit3(node BNode) (uint16, [3]BNode) {
	if node.nbytes() <= PAGE_SIZE {
		// Node within page size, don't split.
		node = node[:PAGE_SIZE]
		return 1, [3]BNode{node}
	}

	left := BNode(make([]byte, 2*PAGE_SIZE)) // Double as it may be split later.
	right := BNode(make([]byte, PAGE_SIZE))
	NodeSplit2(left, right, node)
	if left.nbytes() <= PAGE_SIZE {
		left = left[:PAGE_SIZE]
		return 2, [3]BNode{left, right}
	}

	newLeft := BNode(make([]byte, PAGE_SIZE))
	middle := BNode(make([]byte, PAGE_SIZE))
	NodeSplit2(newLeft, middle, left)
	// newLeft is now guaranteed to not be exceeding page size.
	return 3, [3]BNode{newLeft, middle, right}
}

// Splits the old node into a left and right node. The right node is guaranteed to be within the page size.
// Assumes old node has 2+ keys.
func NodeSplit2(leftNode, rightNode, oldNode BNode) {
	// Greedily guess the number of keys to go into leftNode.
	nleft := oldNode.nkeys() / 2
	nleftBytes := func() uint16 {
		return (BNode{}).estimateBytes(nleft, oldNode.getOffset(nleft))
	}
	for nleftBytes() > PAGE_SIZE {
		nleft--
	}
	// nleft should always be >=1, as worst case is a single large key-value, but max sizes mean that should always fit.

	nrightBytes := func() uint16 {
		// Extra 4B for the additional header needed for the split.
		return oldNode.nbytes() - nleftBytes() + 4
	}
	for nrightBytes() > PAGE_SIZE {
		nleft++
	}
	// nleft should always be <nkeys(), as worst case is a single large key-value, but max sizes mean that should always fit.
	nright := oldNode.nkeys() - nleft

	// Write left node. Note: no guarantee here that it isn't exceeding page size.
	leftNode.SetHeader(oldNode.btype(), nleft)
	NodeAppendRange(leftNode, oldNode, 0, 0, nleft)

	// Write right node. It is guaranteed to not be exceeding page size.
	rightNode.SetHeader(oldNode.btype(), nright)
	NodeAppendRange(leftNode, oldNode, 0, nleft, nright)
}

// Find the last postion that is less than or equal to the key.
func NodeLookupLessThanOrEqual(node BNode, key []byte) uint16 {
	// Could be done using binary search.
	for i := range node.nkeys() {
		cmp := bytes.Compare(node.GetKey(i), key)
		if cmp == 0 {
			// Equal
			return i
		}
		if cmp > 0 {
			// i-th key larger than target key
			return i - 1
		}
	}
	return node.nkeys() - 1
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
