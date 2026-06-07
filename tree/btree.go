package tree

import "bytes"

type BTree struct {
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

func TreeInsert(tree *BTree, node BNode, key, value []byte) BNode {
	newNode := BNode(make([]byte, 2*PAGE_SIZE)) // Double as it may be split later.
	idx := NodeLookupLessThanOrEqual(node, key)
	switch node.btype() {
	case NODE_TYPE_LEAF:
		if bytes.Equal(key, node.GetKey(idx)) {
			// Matching key, update it.
			LeafUpdate(newNode, node, idx, key, value)
		} else {
			// No matching key, idx is the place to insert it.
			LeafInsert(newNode, node, idx, key, value)
		}
	case NODE_TYPE_INTERNAL:
		// Recursively scan down tree to update lead node.
		childPtr := node.getPtr(idx)
		childNode := TreeInsert(tree, tree.get(childPtr), key, value)

		// Split the child node in case it has exceeded the page size and replace.
		nsplit, split := NodeSplit3(childNode)
		tree.del(childPtr)
		NodeReplaceChildN(tree, newNode, node, idx, split[:nsplit]...)
	}
	return newNode
}

func NodeReplaceChildN(tree *BTree, newNode, oldNode BNode, idx uint16, childNodes ...BNode) {
	newNode.SetHeader(NODE_TYPE_INTERNAL, oldNode.nkeys()+uint16(len(childNodes))-1)
	NodeAppendRange(newNode, oldNode, 0, 0, idx)
	for i, childNode := range childNodes {
		NodeAppendKeyValue(newNode, idx+uint16(i), tree.new(childNode), childNode.GetKey(0), nil)
	}
	NodeAppendRange(newNode, oldNode, idx+uint16(len(childNodes)), idx+1, oldNode.nkeys()-idx-1)
}
