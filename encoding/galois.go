// Package encoding provides Galois Field GF(2⁸) arithmetic operations used by
// the Reed–Solomon error correction encoder.
package encoding

const primitivePoly = 0x11D

var (
	gfExp [512]int
	gfLog [256]int
)

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = x
		gfLog[x] = i
		x <<= 1
		if x >= 256 {
			x ^= primitivePoly
		}
	}
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

func gfAdd(a, b int) int {
	return a ^ b
}

func gfMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[gfLog[a]+gfLog[b]]
}

// GeneratorPoly returns the Reed–Solomon generator polynomial of the given degree
// over GF(2⁸) with primitive polynomial 0x11D. The polynomial is computed by
// iteratively multiplying (x - α^i) for i = 0 to degree-1.
// A degree of 0 or less returns the trivial polynomial {1}.
func GeneratorPoly(degree int) []int {
	if degree <= 0 {
		return []int{1}
	}
	g := []int{1, 1}
	for i := 1; i < degree; i++ {
		root := gfExp[i]
		newG := make([]int, len(g)+1)
		for j := 0; j < len(g); j++ {
			newG[j] ^= g[j]
			newG[j+1] ^= gfMul(g[j], root)
		}
		g = newG
	}
	return g
}
