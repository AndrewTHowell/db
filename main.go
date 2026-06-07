package main

import (
	"db/tree"
	"fmt"
	"log"
)

func main() {
	node := tree.BNode(make([]byte, tree.PAGE_SIZE))
	node.SetHeader(tree.NODE_TYPE_LEAF, 2)
	tree.NodeAppendKeyValue(node, 0, 0, []byte("k1"), []byte("hi"))
	tree.NodeAppendKeyValue(node, 1, 0, []byte("k2"), []byte("hello"))

	fmt.Println(stringKeyValue(node, 0))
	fmt.Println(stringKeyValue(node, 1))
}

func stringKeyValue(node tree.BNode, idx int) (string, string) {
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
