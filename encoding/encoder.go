// Package encoding implements the core QR code encoding pipeline including
// mode selection, data bitstream construction, Reed–Solomon error correction,
// mask pattern selection, and module matrix construction.
//
// The primary entry points are [Encode] and [EncodeWithMask], which accept
// raw data bytes and return a fully encoded [QRCode] struct containing the
// module matrix and metadata. Helper functions like [Capacity], [DataCodewords],
// and [MinVersionForData] allow callers to query encoding limits.
//
//	qr, err := encoding.Encode([]byte("hello world"), encoding.ECLevelM, 0)
//	// qr.Modules contains the boolean module matrix
//	// qr.Version, qr.Size, qr.MaskPattern contain metadata
package encoding

import (
	"fmt"
)

// ModeNumeric indicates numeric-only data encoding (digits 0–9).
const ModeNumeric = 1

// ModeAlphanumeric indicates alphanumeric data encoding (0–9, A–Z, and selected symbols).
const ModeAlphanumeric = 2

// ModeByte indicates 8-bit byte data encoding.
const ModeByte = 4

// ModeKanji indicates Shift JIS kanji data encoding.
const ModeKanji = 8

var modeIndicatorBits = map[int][]bool{
	ModeNumeric:      {false, false, false, true},
	ModeAlphanumeric: {false, false, true, false},
	ModeByte:         {false, true, false, false},
	ModeKanji:        {true, false, false, false},
}

// QRCode represents a fully encoded QR code with its module matrix and metadata.
//
// It is the primary output type of [Encode] and [EncodeWithMask].
type QRCode struct {
	// Version is the QR code version (1–40).
	Version int
	// Size is the width and height of the module matrix.
	Size int
	// Modules is the 2D boolean matrix where true represents a dark module.
	Modules [][]bool
	// ECLevel is the error correction level (0=L, 1=M, 2=Q, 3=H).
	ECLevel int
	// MaskPattern is the applied mask pattern (0–7).
	MaskPattern int
	// Data is the original input data.
	Data []byte
	// Metadata holds key-value pairs describing encoding parameters.
	Metadata map[string]string
}

var alphanumericTable = map[byte]int{
	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
	'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14,
	'F': 15, 'G': 16, 'H': 17, 'I': 18, 'J': 19,
	'K': 20, 'L': 21, 'M': 22, 'N': 23, 'O': 24,
	'P': 25, 'Q': 26, 'R': 27, 'S': 28, 'T': 29,
	'U': 30, 'V': 31, 'W': 32, 'X': 33, 'Y': 34,
	'Z': 35, ' ': 36, '$': 37, '%': 38, '*': 39,
	'+': 40, '-': 41, '.': 42, '/': 43, ':': 44,
}

// IsAlphanumeric reports whether b is in the QR alphanumeric character set.
func IsAlphanumeric(b byte) bool {
	_, ok := alphanumericTable[b]
	return ok
}

// IsNumeric reports whether b is a decimal digit.
func IsNumeric(b byte) bool {
	return b >= '0' && b <= '9'
}

// AlphanumericValue returns the numeric code for an alphanumeric character and whether
// the character is valid.
func AlphanumericValue(b byte) (int, bool) {
	val, ok := alphanumericTable[b]
	return val, ok
}

func isKanjiByte1(b byte) bool {
	return (b >= 0x81 && b <= 0x9F) || (b >= 0xE0 && b <= 0xEF)
}

func isKanjiByte2(b byte) bool {
	return (b >= 0x40 && b <= 0x7E) || (b >= 0x80 && b <= 0xFC)
}

// BestEncodingMode returns the most compact encoding mode for the given data.
func BestEncodingMode(data []byte) int {
	if len(data) == 0 {
		return ModeByte
	}
	if len(data)%2 == 0 {
		allKanji := true
		for i := 0; i < len(data); i += 2 {
			if !isKanjiByte1(data[i]) || !isKanjiByte2(data[i+1]) {
				allKanji = false
				break
			}
		}
		if allKanji {
			return ModeKanji
		}
	}
	allNumeric := true
	allAlphanumeric := true
	for _, b := range data {
		if !IsNumeric(b) {
			allNumeric = false
		}
		if !IsAlphanumeric(b) {
			allAlphanumeric = false
		}
		if !allNumeric && !allAlphanumeric {
			return ModeByte
		}
	}
	if allNumeric {
		return ModeNumeric
	}
	if allAlphanumeric {
		return ModeAlphanumeric
	}
	return ModeByte
}

// Encode produces a QR code from data using the specified error correction level
// (0=L, 1=M, 2=Q, 3=H) and version (1–40). A version of 0 triggers automatic
// version selection based on data length. The mask pattern is chosen automatically
// to minimize the penalty score.
//
// Returns an error if ecLevel is out of range, data is empty, or the data
// exceeds the maximum capacity for the given error correction level.
func Encode(data []byte, ecLevel, version int) (*QRCode, error) {
	if ecLevel < 0 || ecLevel > 3 {
		return nil, fmt.Errorf("invalid EC level %d: must be 0-3 (L/M/Q/H)", ecLevel)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("data must not be empty")
	}
	mode := BestEncodingMode(data)
	if version == 0 {
		dataBits := estimateDataBits(data, mode, 10)
		dataBytes := (dataBits + 7) / 8
		var err error
		version, err = MinVersionForData(dataBytes, ecLevel)
		if err != nil {
			return nil, fmt.Errorf("data too large: %w", err)
		}
	} else if version < 1 || version > 40 {
		return nil, fmt.Errorf("invalid version %d: must be 1-40", version)
	}
	vi, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid version/EC level: %w", err)
	}
	totalDataBits := vi.DataCodewords * 8
	dataBits := buildDataBitstream(data, mode, version, totalDataBits)
	dataCodewords := bitsToBytes(dataBits)
	dataBlocks := splitIntoBlocks(dataCodewords, vi)
	ecBlocks := EncodeECBlocks(dataBlocks, vi.ECCodewordsPerBlock)
	interleaved := interleaveBlocks(dataBlocks, ecBlocks, vi)
	interleavedBits := bytesToBits(interleaved)
	matrix := BuildMatrix(version)
	PlaceDataBits(matrix, interleavedBits, version)
	bestMask := BestMaskPattern(matrix, ecLevel, version)
	ApplyMask(matrix, bestMask, version)
	PlaceFormatInfo(matrix, ecLevel, bestMask)
	PlaceVersionInfo(matrix, version)
	size := MatrixSize(version)
	qr := &QRCode{
		Version:     version,
		Size:        size,
		Modules:     matrix,
		ECLevel:     ecLevel,
		MaskPattern: bestMask,
		Data:        data,
		Metadata:    make(map[string]string),
	}
	qr.Metadata["mode"] = modeName(mode)
	qr.Metadata["version"] = fmt.Sprintf("%d", version)
	qr.Metadata["ecLevel"] = ecLevelName(ecLevel)
	qr.Metadata["maskPattern"] = fmt.Sprintf("%d", bestMask)
	qr.Metadata["dataCodewords"] = fmt.Sprintf("%d", vi.DataCodewords)
	qr.Metadata["ecCodewords"] = fmt.Sprintf("%d", vi.DataCodewords+vi.ECBlocks*vi.ECCodewordsPerBlock)
	return qr, nil
}

// EncodeWithMask produces a QR code like [Encode] but allows specifying a
// mask pattern (0–7). A maskPattern of -1 triggers automatic selection using
// penalty scoring, identical to [Encode].
//
// Returns an error if ecLevel or maskPattern is out of range, data is empty,
// or the data exceeds the maximum capacity.
func EncodeWithMask(data []byte, ecLevel, version, maskPattern int) (*QRCode, error) {
	if ecLevel < 0 || ecLevel > 3 {
		return nil, fmt.Errorf("invalid EC level %d: must be 0-3 (L/M/Q/H)", ecLevel)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("data must not be empty")
	}
	if maskPattern < -1 || maskPattern > 7 {
		return nil, fmt.Errorf("invalid mask pattern %d: must be 0-7 or -1 for auto", maskPattern)
	}
	mode := BestEncodingMode(data)
	if version == 0 {
		dataBits := estimateDataBits(data, mode, 10)
		dataBytes := (dataBits + 7) / 8
		var err error
		version, err = MinVersionForData(dataBytes, ecLevel)
		if err != nil {
			return nil, fmt.Errorf("data too large: %w", err)
		}
	} else if version < 1 || version > 40 {
		return nil, fmt.Errorf("invalid version %d: must be 1-40", version)
	}
	vi, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid version/EC level: %w", err)
	}
	totalDataBits := vi.DataCodewords * 8
	dataBits := buildDataBitstream(data, mode, version, totalDataBits)
	dataCodewords := bitsToBytes(dataBits)
	dataBlocks := splitIntoBlocks(dataCodewords, vi)
	ecBlocks := EncodeECBlocks(dataBlocks, vi.ECCodewordsPerBlock)
	interleaved := interleaveBlocks(dataBlocks, ecBlocks, vi)
	interleavedBits := bytesToBits(interleaved)
	matrix := BuildMatrix(version)
	PlaceDataBits(matrix, interleavedBits, version)
	var selectedMask int
	if maskPattern == -1 {
		selectedMask = BestMaskPattern(matrix, ecLevel, version)
	} else {
		selectedMask = maskPattern
	}
	ApplyMask(matrix, selectedMask, version)
	PlaceFormatInfo(matrix, ecLevel, selectedMask)
	PlaceVersionInfo(matrix, version)
	size := MatrixSize(version)
	qr := &QRCode{
		Version:     version,
		Size:        size,
		Modules:     matrix,
		ECLevel:     ecLevel,
		MaskPattern: selectedMask,
		Data:        data,
		Metadata:    make(map[string]string),
	}
	qr.Metadata["mode"] = modeName(mode)
	qr.Metadata["version"] = fmt.Sprintf("%d", version)
	qr.Metadata["ecLevel"] = ecLevelName(ecLevel)
	qr.Metadata["maskPattern"] = fmt.Sprintf("%d", selectedMask)
	qr.Metadata["dataCodewords"] = fmt.Sprintf("%d", vi.DataCodewords)
	qr.Metadata["ecCodewords"] = fmt.Sprintf("%d", vi.DataCodewords+vi.ECBlocks*vi.ECCodewordsPerBlock)
	return qr, nil
}

func estimateDataBits(data []byte, mode, version int) int {
	bits := 4
	ccBits := charCountBits(mode, version)
	bits += ccBits
	switch mode {
	case ModeNumeric:
		bits += len(data) * 10 / 3
	case ModeAlphanumeric:
		bits += len(data) * 11 / 2
	case ModeByte:
		bits += len(data) * 8
	case ModeKanji:
		bits += len(data) / 2 * 13
	}
	bits += 4
	return bits
}

func buildDataBitstream(data []byte, mode, version, totalDataBits int) []bool {
	bits := make([]bool, 0, totalDataBits)
	addModeIndicator(&bits, mode)
	addCharacterCount(&bits, len(data), mode, version)
	addDataBits(&bits, data, mode)
	addTerminatorAndPadding(&bits, totalDataBits)
	return bits
}

func addModeIndicator(bits *[]bool, mode int) {
	indicator := modeIndicatorBits[mode]
	*bits = append(*bits, indicator...)
}

func addCharacterCount(bits *[]bool, count, mode, version int) {
	ccBits := charCountBits(mode, version)
	for i := ccBits - 1; i >= 0; i-- {
		*bits = append(*bits, (count&(1<<uint(i))) != 0)
	}
}

func charCountBits(mode, version int) int {
	switch mode {
	case ModeNumeric:
		if version <= 9 {
			return 10
		}
		return 12
	case ModeAlphanumeric:
		if version <= 9 {
			return 9
		}
		return 11
	case ModeByte:
		if version <= 9 {
			return 8
		}
		return 16
	case ModeKanji:
		if version <= 9 {
			return 8
		}
		return 10
	default:
		return 8
	}
}

func addDataBits(bits *[]bool, data []byte, mode int) {
	switch mode {
	case ModeNumeric:
		encodeNumeric(bits, data)
	case ModeAlphanumeric:
		encodeAlphanumeric(bits, data)
	case ModeByte:
		encodeByte(bits, data)
	case ModeKanji:
		encodeKanji(bits, data)
	}
}

func encodeNumeric(bits *[]bool, data []byte) {
	i := 0
	for i < len(data) {
		remaining := len(data) - i
		if remaining >= 3 { //nolint:gocritic // ifElseChain: numeric encoding has 3 distinct branches by remaining length
			val := int(data[i]-'0')*100 + int(data[i+1]-'0')*10 + int(data[i+2]-'0')
			appendBits(bits, val, 10)
			i += 3
		} else if remaining == 2 {
			val := int(data[i]-'0')*10 + int(data[i+1]-'0')
			appendBits(bits, val, 7)
			i += 2
		} else {
			val := int(data[i] - '0')
			appendBits(bits, val, 4)
			i++
		}
	}
}

func encodeAlphanumeric(bits *[]bool, data []byte) {
	i := 0
	for i < len(data) {
		if i+1 < len(data) {
			val := alphanumericTable[data[i]]*45 + alphanumericTable[data[i+1]]
			appendBits(bits, val, 11)
			i += 2
		} else {
			val := alphanumericTable[data[i]]
			appendBits(bits, val, 6)
			i++
		}
	}
}

func encodeByte(bits *[]bool, data []byte) {
	for _, b := range data {
		appendBits(bits, int(b), 8)
	}
}

func encodeKanji(bits *[]bool, data []byte) {
	for i := 0; i+1 < len(data); i += 2 {
		byte1 := int(data[i])
		byte2 := int(data[i+1])
		code := (byte1<<8 | byte2)
		var temp int
		if code >= 0x8140 && code <= 0x9FFC { //nolint:gocritic // ifElseChain: kanji shift-JIS encoding uses range-based branches
			temp = code - 0x8140
		} else if code >= 0xE040 && code <= 0xEBBF {
			temp = code - 0xC140
		} else {
			continue
		}
		encoded := (temp>>8)*192 + (temp & 0xFF)
		appendBits(bits, encoded, 13)
	}
}

func addTerminatorAndPadding(bits *[]bool, totalDataBits int) {
	currentLen := len(*bits)
	terminatorLen := 4
	if currentLen+terminatorLen > totalDataBits {
		terminatorLen = totalDataBits - currentLen
	}
	for i := 0; i < terminatorLen; i++ {
		*bits = append(*bits, false)
	}
	currentLen += terminatorLen
	byteLen := (currentLen + 7) / 8
	if byteLen*8 > currentLen {
		padding := byteLen*8 - currentLen
		for i := 0; i < padding; i++ {
			*bits = append(*bits, false)
		}
		currentLen = byteLen * 8
	}
	padBytes := []byte{0xEC, 0x11}
	padIndex := 0
	for currentLen < totalDataBits {
		appendBits(bits, int(padBytes[padIndex]), 8)
		padIndex = 1 - padIndex
		currentLen += 8
	}
	if len(*bits) > totalDataBits {
		*bits = (*bits)[:totalDataBits]
	}
}

func appendBits(bits *[]bool, value, numBits int) {
	for i := numBits - 1; i >= 0; i-- {
		*bits = append(*bits, (value&(1<<uint(i))) != 0)
	}
}

func splitIntoBlocks(data []byte, vi *VersionInfo) [][]byte {
	blocks := make([][]byte, 0, vi.ECBlocks)
	offset := 0
	for i := 0; i < vi.Group1Blocks; i++ {
		block := make([]byte, vi.Group1DataCodewords)
		copy(block, data[offset:offset+vi.Group1DataCodewords])
		blocks = append(blocks, block)
		offset += vi.Group1DataCodewords
	}
	for i := 0; i < vi.Group2Blocks; i++ {
		block := make([]byte, vi.Group2DataCodewords)
		copy(block, data[offset:offset+vi.Group2DataCodewords])
		blocks = append(blocks, block)
		offset += vi.Group2DataCodewords
	}
	return blocks
}

func interleaveBlocks(dataBlocks, ecBlocks [][]byte, vi *VersionInfo) []byte {
	totalData := 0
	for _, block := range dataBlocks {
		totalData += len(block)
	}
	totalEC := 0
	for _, block := range ecBlocks {
		totalEC += len(block)
	}
	result := make([]byte, 0, totalData+totalEC)
	maxDataLen := vi.Group1DataCodewords
	if vi.Group2DataCodewords > maxDataLen {
		maxDataLen = vi.Group2DataCodewords
	}
	for i := 0; i < maxDataLen; i++ {
		for _, block := range dataBlocks {
			if i < len(block) {
				result = append(result, block[i])
			}
		}
	}
	for i := 0; i < vi.ECCodewordsPerBlock; i++ {
		for _, block := range ecBlocks {
			if i < len(block) {
				result = append(result, block[i])
			}
		}
	}
	return result
}

// versionRemainderBits is reserved for future API use.
//
//nolint:unused // reserved for future QR code versioning features
func versionRemainderBits(version int) int {
	switch version {
	case 1:
		return 0
	case 2, 3, 4, 5, 6:
		return 7
	case 7, 8, 9, 10, 11, 12, 13:
		return 0
	case 14, 15, 16, 17, 18, 19, 20:
		return 3
	case 21, 22, 23, 24, 25, 26, 27:
		return 4
	case 28, 29, 30, 31, 32, 33, 34:
		return 3
	default:
		return 0
	}
}

func bitsToBytes(bits []bool) []byte {
	bytes := make([]byte, (len(bits)+7)/8)
	for i, bit := range bits {
		if bit {
			bytes[i/8] |= 1 << uint(7-i%8)
		}
	}
	return bytes
}

func bytesToBits(data []byte) []bool {
	bits := make([]bool, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b & (1 << uint(7-j))) != 0
		}
	}
	return bits
}

func modeName(mode int) string {
	switch mode {
	case ModeNumeric:
		return "Numeric"
	case ModeAlphanumeric:
		return "Alphanumeric"
	case ModeByte:
		return "Byte"
	case ModeKanji:
		return "Kanji"
	default:
		return "Unknown"
	}
}

func ecLevelName(level int) string {
	switch level {
	case ECLevelL:
		return "L"
	case ECLevelM:
		return "M"
	case ECLevelQ:
		return "Q"
	case ECLevelH:
		return "H"
	default:
		return "Unknown"
	}
}

// Capacity returns the maximum number of data characters that can be encoded
// in a QR code of the given version, error correction level, and encoding mode.
// The mode must be one of [ModeNumeric], [ModeAlphanumeric], [ModeByte], or
// [ModeKanji]. Returns 0 if version or ecLevel is invalid.
func Capacity(version, ecLevel, mode int) int {
	vi, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return 0
	}
	ccBits := charCountBits(mode, version)
	availableBits := vi.DataCodewords*8 - 4 - ccBits - 4
	if availableBits < 0 {
		return 0
	}
	switch mode {
	case ModeNumeric:
		groups := availableBits / 10
		remainder := availableBits % 10
		count := groups * 3
		if remainder >= 7 {
			count += 2
		} else if remainder >= 4 {
			count++
		}
		return count
	case ModeAlphanumeric:
		groups := availableBits / 11
		remainder := availableBits % 11
		count := groups * 2
		if remainder >= 6 {
			count++
		}
		return count
	case ModeByte:
		return availableBits / 8
	case ModeKanji:
		return availableBits / 13
	default:
		return 0
	}
}

// DataCodewords returns the number of data codewords available for the given
// version and error correction level. Returns 0 if version or ecLevel is invalid.
func DataCodewords(version, ecLevel int) int {
	vi, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return 0
	}
	return vi.DataCodewords
}

// ECCodewords returns the total number of error correction codewords for the
// given version and error correction level (summed across all blocks).
// Returns 0 if version or ecLevel is invalid.
func ECCodewords(version, ecLevel int) int {
	vi, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return 0
	}
	return vi.ECBlocks * vi.ECCodewordsPerBlock
}

// BlockCount returns the number of data blocks in each group for the given
// version and error correction level. Most versions use a single group; larger
// versions split into two groups with different data codeword counts.
// Returns (0, 0) if version or ecLevel is invalid.
func BlockCount(version, ecLevel int) (group1Blocks, group2Blocks int) {
	vi, err := GetVersionInfo(version, ecLevel)
	if err != nil {
		return 0, 0
	}
	return vi.Group1Blocks, vi.Group2Blocks
}
