package encoding

const primitivePoly = 0x11D

var (
	gfExp [512]int
	gfLog [256]int
)

func init() {
	x := 1
	for i := range 255 {
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

func GeneratorPoly(degree int) []int {
	if degree <= 0 {
		return []int{1}
	}
	g := []int{1, 1}
	for i := 1; i < degree; i++ {
		root := gfExp[i]
		newG := make([]int, len(g)+1)
		for j := range g {
			newG[j] ^= g[j]
			newG[j+1] ^= gfMul(g[j], root)
		}
		g = newG
	}
	return g
}
