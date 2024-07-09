package core

import (
	"math/bits"
)

// A bit string is literally a sequence of bits (0s and 1s).
// It is used as a compact way to represent a set of integers.
//
// For example, the bitstring 1010 represents the set {1, 3}.
// The size of the string is 4 bits, and can represent a set of 4 integers.
// Bit strings become efficient to use when the number of integers is large.
// ie. when we have a set of 1000 integers, we can represent it with:
// - naively: 1000 * uint32 = 4000 bytes
// - with a bitstring: 1000 bits = 125 bytes
type Bitstring struct {
	buf []byte
}

func NewBitstring(size int) *Bitstring {
	return &Bitstring{buf: make([]byte, (size / 8) + 1)}
}

// Size returns the number of bits in the bitstring.
func (b *Bitstring) Size() int {
	return len(b.buf) * 8
}

// Count returns the number of bits set in the bitstring.
func (b *Bitstring) Count() int {
	count := 0
	for _, x := range b.buf {
		count += bits.OnesCount8(x)
	}
	return count
}

// SetBit sets the ith bit in the bitstring to 1.
func (b *Bitstring) SetBit(i int) {
	b.buf[i/8] |= 1 << uint(i%8)
}

// SetBit sets the ith bit in the bitstring to 0.
func (b *Bitstring) UnsetBit(i int) {
	b.buf[i/8] &^= 1 << uint(i%8)
}

func (b *Bitstring) IsSet(i int) bool {
	return b.buf[i/8]&(1<<uint(i%8)) != 0
}

// Indices returns a list of all the indices where the bitstring is set.
func (b *Bitstring) Indices() []int {
	var indices []int
	for i := 0; i < len(b.buf)*8; i++ {
		if b.IsSet(i) {
			indices = append(indices, i)
		}
	}
	return indices
}