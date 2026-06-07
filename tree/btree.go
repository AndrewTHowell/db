package tree

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
