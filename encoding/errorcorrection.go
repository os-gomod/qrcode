// Package encoding provides Reed–Solomon error correction encoding over GF(2⁸)
// for QR code data blocks.
package encoding

// EncodeEC computes ecCount error correction bytes for the given data block
// using polynomial long division over GF(2⁸). The returned slice has length
// ecCount. Returns nil if ecCount is zero or data is empty.
func EncodeEC(data []byte, ecCount int) []byte {
	if ecCount <= 0 || len(data) == 0 {
		return nil
	}
	generator := GeneratorPoly(ecCount)
	msgPoly := make([]int, len(data)+ecCount)
	for i := 0; i < len(data); i++ {
		msgPoly[i] = int(data[i])
	}
	for i := 0; i < len(data); i++ {
		if msgPoly[i] == 0 {
			continue
		}
		coef := msgPoly[i]
		for j := 0; j < ecCount; j++ {
			msgPoly[i+j+1] = gfAdd(msgPoly[i+j+1], gfMul(generator[j+1], coef))
		}
	}
	ec := make([]byte, ecCount)
	for i := 0; i < ecCount; i++ {
		ec[i] = byte(msgPoly[len(data)+i])
	}
	return ec
}

// EncodeECBlocks computes error correction bytes for each data block using
// [EncodeEC]. The returned slice is aligned by index with the input blocks.
func EncodeECBlocks(dataBlocks [][]byte, ecPerBlock int) [][]byte {
	ecBlocks := make([][]byte, len(dataBlocks))
	for i, block := range dataBlocks {
		ecBlocks[i] = EncodeEC(block, ecPerBlock)
	}
	return ecBlocks
}
