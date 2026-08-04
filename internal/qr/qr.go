package qr

// qr.go is a QR encoder: byte mode, error-correction level M, versions 1 to 10
// (change a066).
//
// # Why this exists in a trading repository
//
// a065 made turning critical alerts on a single button press, and then handed the
// operator a 26-character random address to type into a phone. The address is
// random because it is the access control; typing it is the friction a065 exists
// to remove; and a phone is exactly where an alert about a position that stopped
// being protected needs to arrive.
//
// # Why it is written rather than imported
//
// This repository takes no dependency it does not need, and a QR encoder is a
// closed problem: the standard fixes every table, the input is one short URL, and
// the output is checkable. What it is NOT is a place to be clever — every table
// below is transcribed from ISO/IEC 18004 and every step is the standard's step in
// the standard's order.
//
// # What is deliberately not supported
//
//   - Modes other than byte. Alphanumeric mode has no lowercase, and the channel
//     is lowercase base32.
//   - Error-correction levels other than M. One level is one set of tables to get
//     right; M is the usual default and recovers about 15%.
//   - Versions above 10. Ten holds 213 bytes at level M, which is far past any
//     ntfy URL including a self-hosted domain. Refusing beyond it is better than
//     carrying thirty more table rows nothing reaches.
//
// # How it is checked
//
// qr_test.go verifies the two BCH computations against the published constants in
// the standard (the format-information strings for level M and the version
// information for versions 7 to 10), checks the Reed-Solomon output by its
// syndromes rather than by re-running the encoder, walks the structure the
// standard specifies, and round-trips the payload back out of the matrix.

import (
	"errors"
	"fmt"
)

// MaxVersion is the largest symbol this encoder builds.
const MaxVersion = 10

// ErrTooLong means the payload does not fit in a version-10 symbol at level M.
var ErrTooLong = errors.New("qr: the payload does not fit in a version-10 symbol")

// Matrix is a finished symbol. Rows are top to bottom, columns left to right, and
// true is a dark module.
type Matrix [][]bool

// Size is the width and height in modules.
func (m Matrix) Size() int { return len(m) }

// blockPlan is one version's data/EC layout at level M, transcribed from the
// standard's error-correction characteristics table.
//
// The four numbers after the EC count are the two block groups: how many blocks
// and how many DATA codewords each carries. A version with one group leaves the
// second pair at zero.
type blockPlan struct {
	total    int // data + EC codewords in the whole symbol
	ecPer    int // EC codewords per block
	g1Blocks int
	g1Data   int
	g2Blocks int
	g2Data   int
}

// plansM is level M for versions 1 to 10, indexed by version.
var plansM = [MaxVersion + 1]blockPlan{
	1:  {26, 10, 1, 16, 0, 0},
	2:  {44, 16, 1, 28, 0, 0},
	3:  {70, 26, 1, 44, 0, 0},
	4:  {100, 18, 2, 32, 0, 0},
	5:  {134, 24, 2, 43, 0, 0},
	6:  {172, 16, 4, 27, 0, 0},
	7:  {196, 18, 4, 31, 0, 0},
	8:  {242, 22, 2, 38, 2, 39},
	9:  {292, 22, 3, 36, 2, 37},
	10: {346, 26, 4, 43, 1, 44},
}

// dataCodewords is how many of a version's codewords carry the message.
func (p blockPlan) dataCodewords() int {
	return p.g1Blocks*p.g1Data + p.g2Blocks*p.g2Data
}

// blocks is the total number of interleaving blocks.
func (p blockPlan) blocks() int { return p.g1Blocks + p.g2Blocks }

// alignmentCentres is the row/column coordinates of alignment-pattern centres,
// per version. Version 1 has none.
var alignmentCentres = [MaxVersion + 1][]int{
	2:  {6, 18},
	3:  {6, 22},
	4:  {6, 26},
	5:  {6, 30},
	6:  {6, 34},
	7:  {6, 22, 38},
	8:  {6, 24, 42},
	9:  {6, 26, 46},
	10: {6, 28, 50},
}

// remainderBits is the padding after the last codeword, per version.
var remainderBits = [MaxVersion + 1]int{
	1: 0, 2: 7, 3: 7, 4: 7, 5: 7, 6: 7, 7: 0, 8: 0, 9: 0, 10: 0,
}

// charCountBits is the length field's width in byte mode: 8 bits up to version 9,
// 16 from version 10.
func charCountBits(version int) int {
	if version >= 10 {
		return 16
	}
	return 8
}

// Encode builds the smallest level-M symbol that holds payload.
func Encode(payload string) (Matrix, error) {
	version, err := smallestVersion(len(payload))
	if err != nil {
		return nil, err
	}
	plan := plansM[version]

	bits := encodeBits(payload, version, plan.dataCodewords())
	interleaved := interleave(bits, plan)

	// Every mask is built and scored; the standard's penalty rules pick one. A
	// wrong choice here still scans — the rules exist to avoid patterns a reader
	// confuses with a finder — so this is the one step where being off by a little
	// is not a correctness failure.
	best := Matrix(nil)
	bestPenalty := -1
	for mask := 0; mask < 8; mask++ {
		candidate := place(version, interleaved, mask)
		if p := penalty(candidate); bestPenalty < 0 || p < bestPenalty {
			best, bestPenalty = candidate, p
		}
	}
	return best, nil
}

// smallestVersion is the first version whose data capacity holds n bytes.
func smallestVersion(n int) (int, error) {
	for version := 1; version <= MaxVersion; version++ {
		capacity := plansM[version].dataCodewords()*8 - 4 - charCountBits(version)
		if n*8 <= capacity {
			return version, nil
		}
	}
	return 0, fmt.Errorf("%w: %d bytes", ErrTooLong, n)
}

// --- the message ------------------------------------------------------------------

// bitWriter accumulates a bit stream as codewords.
type bitWriter struct {
	out  []byte
	bits int // bits used in the last byte
}

func (w *bitWriter) write(value, count int) {
	for i := count - 1; i >= 0; i-- {
		if w.bits == 0 {
			w.out = append(w.out, 0)
		}
		if value&(1<<uint(i)) != 0 {
			w.out[len(w.out)-1] |= 1 << uint(7-w.bits)
		}
		w.bits = (w.bits + 1) % 8
	}
}

// encodeBits builds the data codewords: mode, length, payload, terminator, pad.
func encodeBits(payload string, version, dataCodewords int) []byte {
	w := &bitWriter{}
	w.write(0b0100, 4) // byte mode
	w.write(len(payload), charCountBits(version))
	for i := 0; i < len(payload); i++ {
		w.write(int(payload[i]), 8)
	}

	// Terminator: up to four zero bits, fewer if the capacity ends first.
	used := len(w.out)*8 - (8-w.bits)%8
	if free := dataCodewords*8 - used; free > 0 {
		terminator := 4
		if free < 4 {
			terminator = free
		}
		w.write(0, terminator)
	}
	if w.bits != 0 {
		w.write(0, 8-w.bits)
	}
	// The standard's two pad codewords, alternating, until the block is full.
	for pad := 0; len(w.out) < dataCodewords; pad++ {
		if pad%2 == 0 {
			w.out = append(w.out, 0xEC)
		} else {
			w.out = append(w.out, 0x11)
		}
	}
	return w.out
}

// interleave splits the data into blocks, appends each block's EC codewords, and
// reorders both the standard's way.
func interleave(data []byte, plan blockPlan) []byte {
	dataBlocks := make([][]byte, 0, plan.blocks())
	ecBlocks := make([][]byte, 0, plan.blocks())

	offset := 0
	take := func(count, size int) {
		for i := 0; i < count; i++ {
			block := data[offset : offset+size]
			offset += size
			dataBlocks = append(dataBlocks, block)
			ecBlocks = append(ecBlocks, reedSolomon(block, plan.ecPer))
		}
	}
	take(plan.g1Blocks, plan.g1Data)
	take(plan.g2Blocks, plan.g2Data)

	out := make([]byte, 0, plan.total)
	longest := plan.g1Data
	if plan.g2Data > longest {
		longest = plan.g2Data
	}
	for i := 0; i < longest; i++ {
		for _, block := range dataBlocks {
			if i < len(block) {
				out = append(out, block[i])
			}
		}
	}
	for i := 0; i < plan.ecPer; i++ {
		for _, block := range ecBlocks {
			out = append(out, block[i])
		}
	}
	return out
}

// --- Reed-Solomon over GF(256) ------------------------------------------------------

// The field is GF(2^8) with the QR primitive polynomial x^8+x^4+x^3+x^2+1 (0x11D)
// and generator 2. The two tables below are its logarithms.
var (
	gfExp [512]byte
	gfLog [256]byte
)

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11D
		}
	}
	// The upper half repeats so a product's exponent never needs a modulo.
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

// gfMul multiplies in the field. Zero is special-cased because it has no log.
func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

// generatorPoly is the product of (x - α^i) for i in [0, degree).
func generatorPoly(degree int) []byte {
	poly := []byte{1}
	for i := 0; i < degree; i++ {
		next := make([]byte, len(poly)+1)
		for j, c := range poly {
			next[j] ^= c
			next[j+1] ^= gfMul(c, gfExp[i])
		}
		poly = next
	}
	return poly
}

// reedSolomon returns the `count` error-correction codewords for one block.
func reedSolomon(data []byte, count int) []byte {
	generator := generatorPoly(count)
	remainder := make([]byte, len(data)+count)
	copy(remainder, data)
	for i := 0; i < len(data); i++ {
		lead := remainder[i]
		if lead == 0 {
			continue
		}
		for j, c := range generator {
			remainder[i+j] ^= gfMul(c, lead)
		}
	}
	return remainder[len(data):]
}

// --- BCH, for the two metadata fields ------------------------------------------------

// bch computes the remainder of value shifted left by the generator's degree.
func bch(value, generator, degree int) int {
	remainder := value << uint(degree)
	generatorBits := bitLength(generator)
	for bitLength(remainder) >= generatorBits {
		remainder ^= generator << uint(bitLength(remainder)-generatorBits)
	}
	return remainder
}

func bitLength(v int) int {
	n := 0
	for v != 0 {
		n++
		v >>= 1
	}
	return n
}

// formatInfo is the 15-bit format string for level M and one mask.
//
// Level M is 0b00 in the two high bits, the mask is the three low bits, the BCH
// remainder follows, and the whole thing is XORed with the standard's 0x5412 so a
// blank symbol is not a valid format string.
func formatInfo(mask int) int {
	data := 0b00<<3 | mask
	return ((data << 10) | bch(data, 0b10100110111, 10)) ^ 0x5412
}

// versionInfo is the 18-bit version string, present from version 7.
func versionInfo(version int) int {
	return (version << 12) | bch(version, 0b1111100100101, 12)
}

// --- the module matrix -----------------------------------------------------------

// reserved marks modules that carry function patterns rather than data.
type builder struct {
	m        Matrix
	reserved [][]bool
	size     int
}

func newBuilder(size int) *builder {
	b := &builder{size: size}
	b.m = make(Matrix, size)
	b.reserved = make([][]bool, size)
	for i := range b.m {
		b.m[i] = make([]bool, size)
		b.reserved[i] = make([]bool, size)
	}
	return b
}

func (b *builder) set(x, y int, dark bool) {
	b.m[y][x] = dark
	b.reserved[y][x] = true
}

// place builds the whole symbol for one mask.
func place(version int, codewords []byte, mask int) Matrix {
	size := 17 + 4*version
	b := newBuilder(size)

	b.finders()
	b.timing()
	b.alignment(version)
	// The one module that is always dark and belongs to no pattern.
	b.set(8, size-8, true)
	b.reserveFormat()
	if version >= 7 {
		b.versionBlocks(version)
	}
	b.data(codewords, mask)
	b.format(mask)
	return b.m
}

// finders draws the three finder patterns and their separators.
func (b *builder) finders() {
	for _, origin := range [][2]int{{0, 0}, {b.size - 7, 0}, {0, b.size - 7}} {
		ox, oy := origin[0], origin[1]
		// The separator ring sits one module outside the 7x7 pattern; the loop
		// covers both by walking -1..7 and clamping.
		for dy := -1; dy <= 7; dy++ {
			for dx := -1; dx <= 7; dx++ {
				x, y := ox+dx, oy+dy
				if x < 0 || y < 0 || x >= b.size || y >= b.size {
					continue
				}
				inner := dx >= 0 && dx <= 6 && dy >= 0 && dy <= 6
				dark := false
				if inner {
					edge := dx == 0 || dx == 6 || dy == 0 || dy == 6
					core := dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4
					dark = edge || core
				}
				b.set(x, y, dark)
			}
		}
	}
}

// timing draws the two alternating lines on row 6 and column 6.
func (b *builder) timing() {
	for i := 8; i < b.size-8; i++ {
		dark := i%2 == 0
		b.set(i, 6, dark)
		b.set(6, i, dark)
	}
}

// alignment draws the 5x5 patterns, skipping the three that would sit on a finder.
//
// The three omitted are named by position — first×first, first×last, last×first —
// and not found by asking whether the centre module is already reserved. Those are
// different questions, and the difference is a real defect this code had: from
// version 7 the centre row and column include an interior value whose patterns sit
// ON the timing lines, which are reserved, so a reserved-centre test silently
// dropped two patterns per symbol. Nothing downstream noticed, because the encoder
// and its reader agreed about the smaller symbol; the module count against the
// standard's capacity table is what caught it.
func (b *builder) alignment(version int) {
	centres := alignmentCentres[version]
	if len(centres) == 0 {
		return
	}
	first, last := centres[0], centres[len(centres)-1]
	onFinder := func(cx, cy int) bool {
		return (cx == first && cy == first) ||
			(cx == first && cy == last) ||
			(cx == last && cy == first)
	}
	for _, cy := range centres {
		for _, cx := range centres {
			if onFinder(cx, cy) {
				continue
			}
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					edge := dx == -2 || dx == 2 || dy == -2 || dy == 2
					b.set(cx+dx, cy+dy, edge || (dx == 0 && dy == 0))
				}
			}
		}
	}
}

// reserveFormat marks the format-information modules so data placement skips them.
func (b *builder) reserveFormat() {
	for i := 0; i < 9; i++ {
		if !b.reserved[8][i] {
			b.set(i, 8, false)
		}
		if !b.reserved[i][8] {
			b.set(8, i, false)
		}
	}
	for i := 0; i < 8; i++ {
		if !b.reserved[8][b.size-1-i] {
			b.set(b.size-1-i, 8, false)
		}
		if !b.reserved[b.size-1-i][8] {
			b.set(8, b.size-1-i, false)
		}
	}
}

// versionBlocks writes the two 6x3 version blocks, present from version 7.
func (b *builder) versionBlocks(version int) {
	info := versionInfo(version)
	for i := 0; i < 18; i++ {
		dark := info&(1<<uint(i)) != 0
		x, y := i/3, b.size-11+i%3
		b.set(x, y, dark)
		b.set(y, x, dark)
	}
}

// format writes the two copies of the format string.
func (b *builder) format(mask int) {
	info := formatInfo(mask)
	for i := 0; i < 15; i++ {
		dark := info&(1<<uint(i)) != 0
		// Copy one: down the left column and along the top row, stepping over the
		// timing module at index 6.
		switch {
		case i < 6:
			b.m[i][8] = dark
		case i == 6:
			b.m[7][8] = dark
		case i == 7:
			b.m[8][8] = dark
		case i == 8:
			b.m[8][7] = dark
		default:
			b.m[8][14-i] = dark
		}
		// Copy two: along the bottom-left and the top-right.
		if i < 8 {
			b.m[8][b.size-1-i] = dark
		} else {
			b.m[b.size-15+i][8] = dark
		}
	}
}

// data walks the standard's zigzag and writes the codeword bits, masking as it goes.
func (b *builder) data(codewords []byte, mask int) {
	bit := 0
	upward := true
	for right := b.size - 1; right >= 0; right -= 2 {
		if right == 6 {
			// Column 6 is the vertical timing pattern; the pairs step over it.
			right--
		}
		for step := 0; step < b.size; step++ {
			y := step
			if upward {
				y = b.size - 1 - step
			}
			for dx := 0; dx < 2; dx++ {
				x := right - dx
				if b.reserved[y][x] {
					continue
				}
				dark := false
				if index := bit / 8; index < len(codewords) {
					dark = codewords[index]&(1<<uint(7-bit%8)) != 0
				}
				if maskAt(mask, y, x) {
					dark = !dark
				}
				b.m[y][x] = dark
				bit++
			}
		}
		upward = !upward
	}
}

// maskAt is the standard's eight mask conditions.
func maskAt(mask, row, col int) bool {
	switch mask {
	case 0:
		return (row+col)%2 == 0
	case 1:
		return row%2 == 0
	case 2:
		return col%3 == 0
	case 3:
		return (row+col)%3 == 0
	case 4:
		return (row/2+col/3)%2 == 0
	case 5:
		return (row*col)%2+(row*col)%3 == 0
	case 6:
		return ((row*col)%2+(row*col)%3)%2 == 0
	default:
		return ((row+col)%2+(row*col)%3)%2 == 0
	}
}

// --- the mask penalty --------------------------------------------------------------

// penalty scores a masked symbol by the standard's four rules. Lower is better.
func penalty(m Matrix) int {
	return runPenalty(m) + blockPenalty(m) + finderLikePenalty(m) + balancePenalty(m)
}

// runPenalty is rule 1: five or more same-coloured modules in a line.
func runPenalty(m Matrix) int {
	size := len(m)
	score := 0
	count := func(get func(i int) bool) {
		run, last := 1, get(0)
		for i := 1; i < size; i++ {
			if v := get(i); v == last {
				run++
				continue
			} else {
				last = v
			}
			if run >= 5 {
				score += 3 + run - 5
			}
			run = 1
		}
		if run >= 5 {
			score += 3 + run - 5
		}
	}
	for i := 0; i < size; i++ {
		row, col := i, i
		count(func(j int) bool { return m[row][j] })
		count(func(j int) bool { return m[j][col] })
	}
	return score
}

// blockPenalty is rule 2: every 2x2 block of one colour.
func blockPenalty(m Matrix) int {
	score := 0
	for y := 0; y+1 < len(m); y++ {
		for x := 0; x+1 < len(m); x++ {
			if m[y][x] == m[y][x+1] && m[y][x] == m[y+1][x] && m[y][x] == m[y+1][x+1] {
				score += 3
			}
		}
	}
	return score
}

// finderLikePenalty is rule 3: the 1:1:3:1:1 sequence with four light modules on
// one side, in either orientation, in rows and in columns.
func finderLikePenalty(m Matrix) int {
	size := len(m)
	patterns := [][]bool{
		{true, false, true, true, true, false, true, false, false, false, false},
		{false, false, false, false, true, false, true, true, true, false, true},
	}
	score := 0
	match := func(get func(i int) bool, start int, want []bool) bool {
		for i, w := range want {
			if get(start+i) != w {
				return false
			}
		}
		return true
	}
	for i := 0; i < size; i++ {
		row, col := i, i
		for start := 0; start+11 <= size; start++ {
			for _, want := range patterns {
				if match(func(j int) bool { return m[row][j] }, start, want) {
					score += 40
				}
				if match(func(j int) bool { return m[j][col] }, start, want) {
					score += 40
				}
			}
		}
	}
	return score
}

// balancePenalty is rule 4: how far the dark proportion is from half.
func balancePenalty(m Matrix) int {
	dark, total := 0, 0
	for _, row := range m {
		for _, v := range row {
			total++
			if v {
				dark++
			}
		}
	}
	percent := dark * 100 / total
	deviation := percent - 50
	if deviation < 0 {
		deviation = -deviation
	}
	return deviation / 5 * 10
}
