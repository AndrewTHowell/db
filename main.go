package main

import (
	"db/db"
	"fmt"
	"log"
)

func main() {
	node := db.BNode(make([]byte, db.BTREE_PAGE_SIZE))
	node.SetHeader(db.BTREE_NODE_TYPE_LEAF, 2)
	db.NodeAppendKeyValue(node, 0, 0, []byte("k1"), []byte("hi"))
	db.NodeAppendKeyValue(node, 1, 0, []byte("k2"), []byte("hello"))

	fmt.Println(stringKeyValue(node, 0))
	fmt.Println(stringKeyValue(node, 1))
}

func stringKeyValue(node db.BNode, idx int) (string, string) {
	key, err := node.GetKey(uint16(idx))
	if err != nil {
		log.Panic(err)
	}
	value, err := node.GetValue(uint16(idx))
	if err != nil {
		log.Panic(err)
	}
	return string(key), string(value)
}
