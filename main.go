package main

import "db/db"

func main() {
	node := db.BNode(make([]byte, db.BTREE_PAGE_SIZE))
	node.SetHeader(db.BTREE_NODE_TYPE_LEAF, 2)
	db.NodeAppendKV(node, 0, 0, []byte("k1"), []byte("hi"))
	db.NodeAppendKV(node, 1, 0, []byte("k2"), []byte("hello"))
}
