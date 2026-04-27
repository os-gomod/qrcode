package hash

import (
	"hash/fnv"
)

func Hash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func HashBytes(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

func Combine(h1, h2 uint64) uint64 {
	h1 ^= h2 + 0x9e3779b97f4a7c15 + (h1 << 6) + (h1 >> 2)
	return h1
}
