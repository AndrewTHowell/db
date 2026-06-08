package tree

import (
	"bytes"
	"fmt"
	"log"
)

const (
	NODE_TYPE_INTERNAL = 1
	NODE_TYPE_LEAF     = 2
	PAGE_SIZE          = 4096 // bytes
)

func init() {
	// Worst-case (node with 1 key-value pair):
	// 2B type, 2B nkeys, 1x8B pointers, 1x2B offsets, 1*4B key/value sizes, xB key, yB value.
	nodeSize := 2 + 2 + 8 + 2 + 4 + MAX_KEY_SIZE + MAX_VALUE_SIZE
	if nodeSize > PAGE_SIZE {
		log.Panicf("tree config incompatible, worst-case node %d exceeds page size %d", nodeSize, PAGE_SIZE)
	}
}

type Tree struct {
	// Root pointer.
	root uint64

	// Disk callbacks.

	// Read data from a page number.
	get func(uint64) []byte
	// Allocate a new page number with data.
	new func([]byte) uint64
	// Deallocate a page number.
	del func(uint64)
}

func New(get func(uint64) []byte, new func([]byte) uint64, del func(uint64)) Tree {
	return Tree{
		get: get,
		new: new,
		del: del,
	}
}

// Insert or update a keyed value.
func (t *Tree) Insert(key, value []byte) error {
	if err := (KeyValue{}).validateSize(key, value); err != nil {
		return fmt.Errorf("inserting key-value pair: %w", err)
	}

	if t.root == 0 {
		// Tree is empty, this is the first node.
		root := Node(make([]byte, PAGE_SIZE))
		root.setHeader(NODE_TYPE_LEAF, 2)
		// Add a dummy key. This means tree covers the whole key space, ensuring lookups will always find a containing node.
		nodeAppendKeyValue(root, 0, 0, nil, nil)
		nodeAppendKeyValue(root, 1, 0, key, value)
		t.root = t.new(root)
		return nil
	}

	node := treeInsert(t, t.get(t.root), key, value)

	// Check if node needs to split.
	nsplit, split := nodeSplit3(node)
	t.del(t.root)
	if nsplit > 1 {
		// Root has split, create a new root and add split nodes as children.
		root := Node(make([]byte, PAGE_SIZE))
		root.setHeader(NODE_TYPE_LEAF, nsplit)
		for i, childNode := range split[:nsplit] {
			ptr, key := t.new(childNode), childNode.setKey(0)
			nodeAppendKeyValue(root, uint16(i), ptr, key, nil)
		}
		t.root = t.new(root)
	} else {
		// Root hasn't split.
		t.root = t.new(split[0])
	}
	return nil
}

// Delete a keyed value, returning whether it existed or not.
func (t *Tree) Delete(key []byte) (bool, error) {
	return false, nil
}

func treeInsert(tree *Tree, node Node, key, value []byte) Node {
	newNode := Node(make([]byte, 2*PAGE_SIZE)) // Double as it may be split later.
	idx := nodeLookupLessThanOrEqual(node, key)
	switch node.btype() {
	case NODE_TYPE_LEAF:
		if bytes.Equal(key, node.setKey(idx)) {
			// Matching key, update it.
			leafUpdate(newNode, node, idx, key, value)
		} else {
			// No matching key, idx is the place to insert it.
			leafInsert(newNode, node, idx, key, value)
		}
	case NODE_TYPE_INTERNAL:
		// Recursively scan down tree to update lead node.
		childPtr := node.getPtr(idx)
		childNode := treeInsert(tree, tree.get(childPtr), key, value)

		// Split the child node in case it has exceeded the page size and replace.
		nsplit, split := nodeSplit3(childNode)
		tree.del(childPtr)
		nodeReplaceChildN(tree, newNode, node, idx, split[:nsplit]...)
	}
	return newNode
}

func nodeReplaceChildN(tree *Tree, newNode, oldNode Node, idx uint16, childNodes ...Node) {
	newNode.setHeader(NODE_TYPE_INTERNAL, oldNode.nkeys()+uint16(len(childNodes))-1)
	nodeAppendRange(newNode, oldNode, 0, 0, idx)
	for i, childNode := range childNodes {
		nodeAppendKeyValue(newNode, idx+uint16(i), tree.new(childNode), childNode.setKey(0), nil)
	}
	nodeAppendRange(newNode, oldNode, idx+uint16(len(childNodes)), idx+1, oldNode.nkeys()-idx-1)
}

func leafInsert(newNode, oldNode Node, idx uint16, key, value []byte) {
	newNode.setHeader(NODE_TYPE_LEAF, oldNode.nkeys()+1)
	nodeAppendRange(newNode, oldNode, 0, 0, idx)
	nodeAppendKeyValue(newNode, idx, 0, key, value)
	nodeAppendRange(newNode, oldNode, idx+1, idx, oldNode.nkeys()-idx)
}

func leafUpdate(newNode, oldNode Node, idx uint16, key, value []byte) {
	newNode.setHeader(NODE_TYPE_LEAF, oldNode.nkeys())
	nodeAppendRange(newNode, oldNode, 0, 0, idx)
	nodeAppendKeyValue(newNode, idx, 0, key, value)
	nodeAppendRange(newNode, oldNode, idx+1, idx+1, oldNode.nkeys()-idx-1)
}

func nodeSplit3(node Node) (uint16, [3]Node) {
	if node.nbytes() <= PAGE_SIZE {
		// Node within page size, don't split.
		node = node[:PAGE_SIZE]
		return 1, [3]Node{node}
	}

	left := Node(make([]byte, 2*PAGE_SIZE)) // Double as it may be split later.
	right := Node(make([]byte, PAGE_SIZE))
	nodeSplit2(left, right, node)
	if left.nbytes() <= PAGE_SIZE {
		left = left[:PAGE_SIZE]
		return 2, [3]Node{left, right}
	}

	newLeft := Node(make([]byte, PAGE_SIZE))
	middle := Node(make([]byte, PAGE_SIZE))
	nodeSplit2(newLeft, middle, left)
	// newLeft is now guaranteed to not be exceeding page size.
	return 3, [3]Node{newLeft, middle, right}
}

// Splits the old node into a left and right node. The right node is guaranteed to be within the page size.
// Assumes old node has 2+ keys.
func nodeSplit2(leftNode, rightNode, oldNode Node) {
	// Greedily guess the number of keys to go into leftNode.
	nleft := oldNode.nkeys() / 2
	nleftBytes := func() uint16 {
		return (Node{}).estimateBytes(nleft, oldNode.getOffset(nleft))
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
	leftNode.setHeader(oldNode.btype(), nleft)
	nodeAppendRange(leftNode, oldNode, 0, 0, nleft)

	// Write right node. It is guaranteed to not be exceeding page size.
	rightNode.setHeader(oldNode.btype(), nright)
	nodeAppendRange(leftNode, oldNode, 0, nleft, nright)
}

// Find the last postion that is less than or equal to the key.
func nodeLookupLessThanOrEqual(node Node, key []byte) uint16 {
	// Could be done using binary search.
	for i := range node.nkeys() {
		cmp := bytes.Compare(node.setKey(i), key)
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

func nodeAppendRange(newNode, oldNode Node, dstNew, srcOld, n uint16) {
	for i := range n {
		dst, src := dstNew+i, srcOld+i
		nodeAppendKeyValue(newNode, dst, oldNode.getPtr(src), oldNode.setKey(src), oldNode.setValue(src))
	}
}

func nodeAppendKeyValue(node Node, idx uint16, ptr uint64, key, value []byte) {
	node.setPtr(idx, ptr)
	keyValue := newKeyValue(node[node.keyValuePosition(idx):], key, value)
	// Set offset for next key.
	node.setOffset(idx+1, node.getOffset(idx)+uint16(len(keyValue)))
}
