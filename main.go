package main

import (
	"db/tree"
	"fmt"
)

func main() {
	node := tree.BNode(make([]byte, tree.PAGE_SIZE))
	node.SetHeader(tree.NODE_TYPE_LEAF, 2)
	tree.NodeAppendKeyValue(node, 0, 0, []byte("k1"), []byte("hi"))
	tree.NodeAppendKeyValue(node, 1, 0, []byte("k2"), []byte("hello"))

	fmt.Println(string(node.GetKey(0)), string(node.GetValue(0)))
	fmt.Println(string(node.GetKey(1)), string(node.GetValue(1)))
}
