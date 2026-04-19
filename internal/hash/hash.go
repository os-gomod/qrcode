// Package hash provides FNV-1a hashing utilities for keys and byte slices.
package hash

import (
	"hash/fnv"
)

// Hash returns the FNV-1a 64-bit hash of the given string key.
func Hash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

// HashBytes returns the FNV-1a hash of the given byte slice.
//
//nolint:revive // stutter: HashBytes is the canonical name for this function
func HashBytes(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

// Combine combines two hash values into one using a boost-style hash
// combining function. This is useful for building composite hash keys.
func Combine(h1, h2 uint64) uint64 {
	h1 ^= h2 + 0x9e3779b97f4a7c15 + (h1 << 6) + (h1 >> 2)
	return h1
}
