package main

import (
	"db/tree"
	"fmt"
	"log"
	"unsafe"
)

func main() {
	pages := map[uint64]tree.Node{} // In-memory pages.

	t := tree.New(
		func(ptr uint64) []byte {
			node, ok := pages[ptr]
			if !ok {
				log.Panic("page doesn't exist")
			}
			return node
		},
		func(data []byte) uint64 {
			ptr := uint64(uintptr(unsafe.Pointer(&data[0])))
			_, ok := pages[ptr]
			if ok {
				log.Panic("page already exists")
			}
			pages[ptr] = data
			return ptr
		},
		func(ptr uint64) {
			_, ok := pages[ptr]
			if !ok {
				log.Panic("page doesn't exist")
			}
			delete(pages, ptr)
		},
	)
	fmt.Println(t.Insert([]byte("k1"), []byte("hi")))
	fmt.Println(t.Insert([]byte("k2"), []byte("hello")))

	for _, page := range pages {
		fmt.Println(string(page))
	}
}
