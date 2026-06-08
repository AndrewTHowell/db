package tree

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	MAX_KEY_SIZE   = 1000 // bytes
	MAX_VALUE_SIZE = 3000 // bytes
)

type ErrKeyExceedsMaxSize struct{}

func (ErrKeyExceedsMaxSize) Error() string {
	return fmt.Sprintf("key exceeds max size %d", MAX_KEY_SIZE)
}

type ErrValueExceedsMaxSize struct{}

func (ErrValueExceedsMaxSize) Error() string {
	return fmt.Sprintf("value exceeds max size %d", MAX_VALUE_SIZE)
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

func (KeyValue) validateSize(key, value []byte) error {
	var err error
	if len(key) > MAX_KEY_SIZE {
		err = errors.Join(err, ErrKeyExceedsMaxSize{})
	}
	if len(value) > MAX_VALUE_SIZE {
		err = errors.Join(err, ErrValueExceedsMaxSize{})
	}
	return err
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
