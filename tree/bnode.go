package tree

import (
	"encoding/binary"
)

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

func (node BNode) nbytes() uint16 {
	// Number of bytes is equivalent to the byte position of the next key-value.
	return node.keyValuePosition(node.nkeys())
}

func (node BNode) SetHeader(btype, nkeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], btype)
	binary.LittleEndian.PutUint16(node[2:4], nkeys)
}

// Assumes idx has been validated to be within valid range.
func (node BNode) getPtr(idx uint16) uint64 {
	// Skip 4B header. Go to Nth ptr with each ptr being 8B.
	return binary.LittleEndian.Uint64(node[4+(idx*8):])
}

// Assumes idx has been validated to be within valid range.
func (node BNode) setPtr(idx uint16, value uint64) {
	// Skip 4B header. Go to Nth ptr with each ptr being 8B.
	binary.LittleEndian.PutUint64(node[4+(idx*8):], value)
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

// Assumes idx has been validated to be within valid range.
func (node BNode) GetKey(idx uint16) []byte {
	return KeyValue(node[node.keyValuePosition(idx):node.keyValuePosition(idx+1)]).getKey()
}

// Assumes idx has been validated to be within valid range.
func (node BNode) GetValue(idx uint16) []byte {
	return KeyValue(node[node.keyValuePosition(idx):node.keyValuePosition(idx+1)]).getValue()
}
