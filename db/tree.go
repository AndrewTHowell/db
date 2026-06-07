package db

import (
	"encoding/binary"
	"fmt"
	"log"
)

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

type ErrIndex string

func (e ErrIndex) Error() string {
	return fmt.Sprintf("index error: %s", string(e))
}

func NodeAppendKeyValue(node BNode, idx uint16, ptr uint64, key, value []byte) error {
	if idx >= node.nkeys() {
		return ErrIndex("list index out of range")
	}
	node.setPtr(idx, ptr)
	keyValue := newKeyValue(node[node.keyValuePosition(idx):], key, value)
	// Set offset for next key.
	node.setOffset(idx+1, node.getOffset(idx)+uint16(len(keyValue)))
	return nil
}

// Binary format - node:
// |    header    |                        data                   |
// | type | nkeys |  pointers  |  offsets   | key-values | unused |
// |  2B  |   2B  | nkeys × 8B | nkeys × 2B |     ...    |        |
type BNode []byte

func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

func (node BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

func (node BNode) SetHeader(btype, nkeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], btype)
	binary.LittleEndian.PutUint16(node[2:4], nkeys)
}

func (node BNode) getPtr(idx uint16) (uint64, error) {
	if idx >= node.nkeys() {
		return 0, ErrIndex("list index out of range")
	}
	// Skip 4B header. Go to Nth ptr with each ptr being 8B.
	return binary.LittleEndian.Uint64(node[4+(idx*8):]), nil
}

func (node BNode) setPtr(idx uint16, value uint64) error {
	if idx >= node.nkeys() {
		return ErrIndex("list index out of range")
	}
	// Skip 4B header. Go to Nth ptr with each ptr being 8B.
	binary.LittleEndian.PutUint64(node[4+(idx*8):], value)
	return nil
}

// Assumes idx has been validated to be within valid range.
func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		// First offset isn't stored, it's always zero.
		return 0
	}
	// Skip 4B header and nkeys*8B pointers.
	// Go to N-1th offset (as 0th isn't stored) with each offset being 2B.
	return binary.LittleEndian.Uint16(node[4+(node.nkeys()*8)+((idx-1)*2):])
}

// Assumes idx has been validated to be within valid range.
func (node BNode) setOffset(idx, offset uint16) {
	if idx == 0 {
		// First offset isn't stored, it's always zero.
		return
	}
	// Skip 4B header and nkeys*8B pointers.
	// Go to N-1th offset (as 0th isn't stored) with each offset being 2B.
	binary.LittleEndian.PutUint16(node[4+(node.nkeys()*8)+((idx-1)*2):], offset)
}

// Assumes idx has been validated to be within valid range.
func (node BNode) keyValuePosition(idx uint16) uint16 {
	// Skip 4B header, nkeys*8B pointers, and nkeys*2B offsets. Go to offset.
	return 4 + (node.nkeys() * 8) + (node.nkeys() * 2) + node.getOffset(idx)
}

func (node BNode) GetKey(idx uint16) ([]byte, error) {
	if idx >= node.nkeys() {
		return nil, ErrIndex("list index out of range")
	}
	return KeyValue(node[node.keyValuePosition(idx):node.keyValuePosition(idx+1)]).getKey(), nil
}

func (node BNode) GetValue(idx uint16) ([]byte, error) {
	if idx >= node.nkeys() {
		return nil, ErrIndex("list index out of range")
	}
	return KeyValue(node[node.keyValuePosition(idx):node.keyValuePosition(idx+1)]).getValue(), nil
}

// Binary format:
// |        header       |   data    |
// | key_size | val_size | key | val |
// |    2B    |    2B    | ... | ... |
type KeyValue []byte

func newKeyValue(data, key, value []byte) KeyValue {
	binary.LittleEndian.PutUint16(data[0:], uint16(len(key)))
	binary.LittleEndian.PutUint16(data[2:], uint16(len(value)))
	// Skip 4B header.
	copy(data[4:], key)
	// Skip 4B header and key.
	copy(data[4+uint16(len(key)):], value)
	// Return slice up to value.
	return KeyValue(data[:4+uint16(len(key)+len(value))])
}

func (kv KeyValue) keySize() uint16 {
	return binary.LittleEndian.Uint16(kv[0:2])
}

func (kv KeyValue) valueSize() uint16 {
	return binary.LittleEndian.Uint16(kv[2:4])
}

func (kv KeyValue) getKey() []byte {
	// Skip 4B header.
	return kv[4 : 4+kv.keySize()]
}

func (kv KeyValue) getValue() []byte {
	// Skip 4B header and key.
	return kv[4+kv.keySize():]
}
