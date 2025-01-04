package core

import (
	"math/bits"
)

// A bit string is a fixed-length string of bits (0s and 1s) used to compactly represent a set of integers. Each bit at index `i` represents the membership of integer `i` in the set.
//
// For example, the bitstring 0101 represents the set {1, 3}.
// The size of the string is 4 bits, and can represent a set of 4 integers.
// Bit sets become efficient to use when the count of integers is large.
// ie. when we have a set of 1000 integers, we can represent it with:
//   - naively as:  1000 * uint32 = 4000 bytes
//   - with a bitstring: 1000 bits = 125 bytes
type Bitset []byte

func NewBitset(size int) *Bitset {
	buf := make([]byte, (size/8)+1)
	return (*Bitset)(&buf)
}

func NewBitsetFromBuffer(buf []byte) *Bitset {
	return (*Bitset)(&buf)
}

// Size returns the number of integers countable in the bit set.
func (b *Bitset) Size() int {
	return len(*b) * 8
}

// Count returns the number of bits set in the bitstring.
func (b *Bitset) Count() int {
	count := 0
	for _, x := range *b {
		count += bits.OnesCount8(x)
	}
	return count
}

// SetBit sets the ith bit in the bitstring to 1.
func (b *Bitset) Insert(i int) {
	(*b)[i/8] |= 1 << uint(i%8)
}

// SetBit sets the ith bit in the bitstring to 0.
func (b *Bitset) Remove(i int) {
	(*b)[i/8] &^= 1 << uint(i%8)
}

func (b *Bitset) Contains(i int) bool {
	return (*b)[i/8]&(1<<uint(i%8)) != 0
}

// Indices returns a list of all the indices where the bitstring is set.
func (b *Bitset) Indices() []int {
	var indices []int
	for i := 0; i < len(*b)*8; i++ {
		if b.Contains(i) {
			indices = append(indices, i)
		}
	}
	return indices
}

// Ranges returns a list of ranges where the bitstring is set.
// Useful for printing.
func (b *Bitset) Ranges() [][2]int {
	return findOnesRanges(*b)
}

func findOnesRanges(data []byte) [][2]int {
	var ranges [][2]int
	start := -1

	for i := 0; i < len(data)*8; i++ {
		bit := data[i/8]&(1<<uint(i%8)) != 0
		if bit {
			if start == -1 {
				start = i
			}
		} else {
			if start != -1 {
				ranges = append(ranges, [2]int{start, i - 1})
				start = -1
			}
		}
	}

	if start != -1 {
		ranges = append(ranges, [2]int{start, len(data)*8 - 1})
	}

	return ranges
}
