package encoding

func EncodeEC(data []byte, ecCount int) []byte {
	if ecCount <= 0 || len(data) == 0 {
		return nil
	}
	generator := GeneratorPoly(ecCount)
	msgPoly := make([]int, len(data)+ecCount)
	for i := range data {
		msgPoly[i] = int(data[i])
	}
	for i := range data {
		if msgPoly[i] == 0 {
			continue
		}
		coef := msgPoly[i]
		for j := range ecCount {
			msgPoly[i+j+1] = gfAdd(msgPoly[i+j+1], gfMul(generator[j+1], coef))
		}
	}
	ec := make([]byte, ecCount)
	for i := range ecCount {
		ec[i] = byte(msgPoly[len(data)+i])
	}
	return ec
}

func EncodeECBlocks(dataBlocks [][]byte, ecPerBlock int) [][]byte {
	ecBlocks := make([][]byte, len(dataBlocks))
	for i, block := range dataBlocks {
		ecBlocks[i] = EncodeEC(block, ecPerBlock)
	}
	return ecBlocks
}
