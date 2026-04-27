package encoding

import (
	"fmt"
)

const (
	ECLevelL = 0
	ECLevelM = 1
	ECLevelQ = 2
	ECLevelH = 3
)

var ecLevelFormatBits = [4]int{1, 0, 3, 2}

type VersionInfo struct {
	Version             int
	TotalCodewords      int
	DataCodewords       int
	ECBlocks            int
	ECCodewordsPerBlock int
	NumGroups           int
	Group1Blocks        int
	Group1DataCodewords int
	Group2Blocks        int
	Group2DataCodewords int
}
type ecLevelParams struct {
	ecCodewordsPerBlock int
	group1Blocks        int
	group1DataCW        int
	group2Blocks        int
	group2DataCW        int
}

var versionData [40]struct {
	totalCW int
	levels  [4]ecLevelParams
}

//nolint:funlen,cyclop // version data initialization table
func init() {
	raw := [][21]int{
		{
			26,
			7, 1, 19, 0, 0,
			10, 1, 16, 0, 0,
			13, 1, 13, 0, 0,
			17, 1, 9, 0, 0,
		},
		{
			44,
			10, 1, 34, 0, 0,
			16, 1, 28, 0, 0,
			22, 1, 22, 0, 0,
			28, 1, 16, 0, 0,
		},
		{
			70,
			15, 1, 55, 0, 0,
			26, 1, 44, 0, 0,
			18, 2, 17, 0, 0,
			22, 2, 13, 0, 0,
		},
		{
			100,
			20, 1, 80, 0, 0,
			18, 2, 32, 0, 0,
			26, 2, 24, 0, 0,
			16, 4, 9, 0, 0,
		},
		{
			134,
			26, 1, 108, 0, 0,
			24, 2, 43, 0, 0,
			18, 2, 15, 2, 16,
			22, 2, 11, 2, 12,
		},
		{
			172,
			18, 2, 68, 0, 0,
			16, 4, 27, 0, 0,
			24, 4, 19, 0, 0,
			28, 4, 15, 0, 0,
		},
		{
			196,
			20, 2, 78, 0, 0,
			18, 4, 31, 0, 0,
			18, 2, 14, 4, 15,
			26, 4, 13, 1, 14,
		},
		{
			242,
			24, 2, 97, 0, 0,
			22, 2, 38, 2, 39,
			22, 4, 18, 2, 19,
			26, 4, 14, 2, 15,
		},
		{
			292,
			30, 2, 116, 0, 0,
			22, 3, 36, 2, 37,
			20, 4, 16, 4, 17,
			24, 4, 12, 4, 13,
		},
		{
			346,
			18, 2, 68, 2, 69,
			26, 4, 43, 1, 44,
			24, 6, 19, 2, 20,
			28, 6, 15, 2, 16,
		},
		{
			404,
			20, 4, 81, 0, 0,
			30, 1, 50, 4, 51,
			28, 4, 22, 4, 23,
			24, 3, 12, 8, 13,
		},
		{
			466,
			24, 2, 92, 2, 93,
			22, 6, 36, 2, 37,
			26, 4, 20, 6, 21,
			28, 7, 14, 4, 15,
		},
		{
			532,
			26, 4, 107, 0, 0,
			22, 8, 37, 1, 38,
			24, 8, 20, 4, 21,
			22, 12, 11, 4, 12,
		},
		{
			581,
			30, 3, 115, 1, 116,
			24, 4, 40, 5, 41,
			20, 11, 16, 5, 17,
			24, 11, 12, 5, 13,
		},
		{
			655,
			22, 5, 87, 1, 88,
			24, 5, 41, 5, 42,
			30, 5, 24, 7, 25,
			24, 11, 12, 7, 13,
		},
		{
			733,
			24, 5, 98, 1, 99,
			28, 7, 45, 3, 46,
			24, 15, 19, 2, 20,
			30, 3, 15, 13, 16,
		},
		{
			815,
			28, 1, 107, 5, 108,
			28, 10, 46, 1, 47,
			28, 1, 22, 15, 23,
			28, 2, 14, 17, 15,
		},
		{
			901,
			30, 5, 120, 1, 121,
			26, 9, 43, 4, 44,
			28, 17, 22, 1, 23,
			28, 2, 14, 19, 15,
		},
		{
			991,
			28, 3, 113, 4, 114,
			26, 3, 44, 11, 45,
			26, 17, 21, 4, 22,
			26, 9, 13, 16, 14,
		},
		{
			1085,
			28, 3, 107, 5, 108,
			26, 3, 41, 13, 42,
			30, 15, 24, 5, 25,
			28, 15, 15, 10, 16,
		},
		{
			1156,
			28, 4, 116, 4, 117,
			26, 17, 42, 0, 0,
			28, 17, 22, 6, 23,
			30, 19, 16, 6, 17,
		},
		{
			1258,
			28, 2, 111, 7, 112,
			28, 17, 46, 0, 0,
			30, 7, 24, 16, 25,
			24, 34, 13, 0, 0,
		},
		{
			1364,
			30, 4, 121, 5, 122,
			28, 4, 47, 14, 48,
			30, 11, 24, 14, 25,
			30, 16, 15, 14, 16,
		},
		{
			1474,
			30, 6, 117, 4, 118,
			28, 6, 45, 14, 46,
			30, 11, 24, 16, 25,
			30, 30, 16, 2, 17,
		},
		{
			1588,
			26, 8, 106, 4, 107,
			28, 8, 47, 13, 48,
			30, 7, 24, 22, 25,
			30, 22, 15, 13, 16,
		},
		{
			1706,
			28, 10, 114, 2, 115,
			28, 19, 46, 4, 47,
			28, 28, 22, 6, 23,
			30, 33, 16, 4, 17,
		},
		{
			1828,
			30, 8, 122, 4, 123,
			28, 22, 45, 3, 46,
			30, 8, 23, 26, 24,
			30, 12, 15, 28, 16,
		},
		{
			1921,
			30, 3, 117, 10, 118,
			28, 3, 45, 23, 46,
			30, 4, 24, 31, 25,
			30, 11, 15, 31, 16,
		},
		{
			2051,
			30, 7, 116, 7, 117,
			28, 21, 45, 7, 46,
			30, 1, 23, 37, 24,
			30, 19, 15, 26, 16,
		},
		{
			2185,
			30, 5, 115, 10, 116,
			28, 19, 47, 10, 48,
			30, 15, 24, 25, 25,
			30, 23, 15, 25, 16,
		},
		{
			2323,
			30, 13, 115, 3, 116,
			28, 2, 46, 29, 47,
			30, 42, 24, 1, 25,
			30, 23, 15, 28, 16,
		},
		{
			2465,
			30, 17, 115, 0, 0,
			28, 10, 46, 23, 47,
			30, 10, 24, 35, 25,
			30, 19, 15, 35, 16,
		},
		{
			2611,
			30, 17, 115, 1, 116,
			28, 14, 46, 21, 47,
			30, 29, 24, 19, 25,
			30, 11, 15, 46, 16,
		},
		{
			2761,
			30, 13, 115, 6, 116,
			28, 14, 46, 23, 47,
			30, 44, 24, 7, 25,
			30, 59, 16, 1, 17,
		},
		{
			2876,
			30, 12, 121, 7, 122,
			28, 12, 47, 26, 48,
			30, 39, 24, 14, 25,
			30, 22, 15, 41, 16,
		},
		{
			3034,
			30, 6, 121, 14, 122,
			28, 6, 47, 34, 48,
			30, 46, 24, 10, 25,
			30, 2, 15, 64, 16,
		},
		{
			3196,
			30, 17, 122, 4, 123,
			28, 29, 46, 14, 47,
			30, 49, 24, 10, 25,
			30, 24, 15, 46, 16,
		},
		{
			3362,
			30, 4, 122, 18, 123,
			28, 13, 46, 32, 47,
			30, 48, 24, 14, 25,
			30, 42, 15, 32, 16,
		},
		{
			3532,
			30, 20, 117, 4, 118,
			28, 40, 47, 7, 48,
			30, 43, 24, 22, 25,
			30, 10, 15, 67, 16,
		},
		{
			3706,
			30, 19, 118, 6, 119,
			28, 18, 47, 31, 48,
			30, 34, 24, 34, 25,
			30, 20, 15, 61, 16,
		},
	}
	for v := range 40 {
		r := raw[v]
		versionData[v].totalCW = r[0]
		for lvl := range 4 {
			base := 1 + lvl*5
			versionData[v].levels[lvl] = ecLevelParams{
				ecCodewordsPerBlock: r[base+0],
				group1Blocks:        r[base+1],
				group1DataCW:        r[base+2],
				group2Blocks:        r[base+3],
				group2DataCW:        r[base+4],
			}
		}
	}
}

func GetVersionInfo(version, ecLevel int) (*VersionInfo, error) {
	if version < 1 || version > 40 {
		return nil, fmt.Errorf("invalid version %d: must be 1-40", version)
	}
	if ecLevel < 0 || ecLevel > 3 {
		return nil, fmt.Errorf("invalid EC level %d: must be 0-3 (L/M/Q/H)", ecLevel)
	}
	vd := &versionData[version-1]
	params := vd.levels[ecLevel]
	info := &VersionInfo{
		Version:             version,
		TotalCodewords:      vd.totalCW,
		ECCodewordsPerBlock: params.ecCodewordsPerBlock,
		Group1Blocks:        params.group1Blocks,
		Group1DataCodewords: params.group1DataCW,
		Group2Blocks:        params.group2Blocks,
		Group2DataCodewords: params.group2DataCW,
	}
	info.ECBlocks = params.group1Blocks + params.group2Blocks
	info.DataCodewords = params.group1Blocks*params.group1DataCW + params.group2Blocks*params.group2DataCW
	if params.group2Blocks > 0 {
		info.NumGroups = 2
	} else {
		info.NumGroups = 1
	}
	return info, nil
}

func DataCapacity(version, ecLevel int) int {
	info, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return 0
	}
	return info.DataCodewords
}

func MinVersionForData(dataLen, ecLevel int) (int, error) {
	if ecLevel < 0 || ecLevel > 3 {
		return 0, fmt.Errorf("invalid EC level %d: must be 0-3 (L/M/Q/H)", ecLevel)
	}
	if dataLen < 0 {
		return 0, fmt.Errorf("data length %d must be non-negative", dataLen)
	}
	for v := 1; v <= 40; v++ {
		capacity := DataCapacity(v, ecLevel)
		if dataLen <= capacity {
			return v, nil
		}
	}
	return 0, fmt.Errorf("data length %d exceeds maximum capacity for EC level %d (version 40 holds %d bytes)",
		dataLen, ecLevel, DataCapacity(40, ecLevel))
}
