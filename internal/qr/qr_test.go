package qr

// qr_test.go checks the encoder four ways, and only one of them re-uses the
// encoder's own idea of what is correct.
//
//	① published constants — the format strings and version strings in the
//	  standard are compared against what the two BCH computations produce. These
//	  are external reference values, not this code's output.
//	② syndromes — the Reed-Solomon codewords are verified by evaluating the
//	  codeword polynomial at the generator's roots, which must be zero. That is a
//	  different computation from the one that produced them.
//	③ structure — finder patterns, separators, timing lines, alignment centres,
//	  the dark module and the reserved areas are walked against the standard.
//	④ round trip — the payload is read back out of the finished matrix through
//	  the format string, the mask and the de-interleaver.
//
// ④ shares the placement order with the encoder and would not catch a symbol
// that is self-consistently wrong about the zigzag. ①②③ are what stand behind it.

import (
	"strings"
	"testing"
)

// --- ① published constants ---------------------------------------------------------

// TestTheFormatStringsMatchTheStandard.
//
// These fifteen-bit values are printed in ISO/IEC 18004's format-information
// table for error-correction level M, one per mask. They are the reference; the
// BCH implementation is what is being checked.
func TestTheFormatStringsMatchTheStandard(t *testing.T) {
	want := [8]int{0x5412, 0x5125, 0x5E7C, 0x5B4B, 0x45F9, 0x40CE, 0x4F97, 0x4AA0}
	for mask, expected := range want {
		if got := formatInfo(mask); got != expected {
			t.Errorf("formatInfo(%d) = %015b, want %015b", mask, got, expected)
		}
	}
}

// TestTheVersionStringsMatchTheStandard.
//
// Version information appears from version 7. These four are the standard's
// values for the versions this encoder can reach.
func TestTheVersionStringsMatchTheStandard(t *testing.T) {
	want := map[int]int{7: 0x07C94, 8: 0x085BC, 9: 0x09A99, 10: 0x0A4D3}
	for version, expected := range want {
		if got := versionInfo(version); got != expected {
			t.Errorf("versionInfo(%d) = %018b, want %018b", version, got, expected)
		}
	}
}

// --- ② Reed-Solomon by its syndromes -------------------------------------------------

// TestTheErrorCorrectionCodewordsHaveZeroSyndromes.
//
// A Reed-Solomon codeword is a polynomial with the generator's roots as its own.
// Evaluating it at α^0 … α^(n-1) must give zero at every one. This never calls
// reedSolomon's arithmetic in the same direction, so a generator polynomial that
// was built wrong shows up here rather than agreeing with itself.
func TestTheErrorCorrectionCodewordsHaveZeroSyndromes(t *testing.T) {
	for _, count := range []int{10, 16, 18, 22, 24, 26} {
		data := make([]byte, 20)
		for i := range data {
			data[i] = byte(i*7 + 3)
		}
		codeword := append(append([]byte{}, data...), reedSolomon(data, count)...)
		for root := 0; root < count; root++ {
			// Horner's method at α^root.
			acc := byte(0)
			for _, c := range codeword {
				acc = gfMul(acc, gfExp[root]) ^ c
			}
			if acc != 0 {
				t.Errorf("ec=%d: syndrome at α^%d is %d, want 0", count, root, acc)
			}
		}
	}
}

// TestTheFieldIsTheOneTheStandardNames.
//
// GF(256) with primitive polynomial 0x11D and generator 2. Two spot values and
// the cycle length are enough to tell it from a neighbouring field.
func TestTheFieldIsTheOneTheStandardNames(t *testing.T) {
	if gfExp[8] != 0x1D {
		t.Errorf("α^8 = %#x, want 0x1D — that is what x^8 = x^4+x^3+x^2+1 means", gfExp[8])
	}
	if gfExp[255] != gfExp[0] {
		t.Error("the exponent table does not wrap at 255; the multiplicative order is wrong")
	}
	seen := map[byte]bool{}
	for i := 0; i < 255; i++ {
		if seen[gfExp[i]] {
			t.Fatalf("α^%d repeats a value; 2 is not a generator of this field", i)
		}
		seen[gfExp[i]] = true
	}
}

// --- ③ structure --------------------------------------------------------------------

// TestTheSymbolHasTheStructureTheStandardRequires.
func TestTheSymbolHasTheStructureTheStandardRequires(t *testing.T) {
	for _, payload := range []string{
		"https://ntfy.sh/tossos-mnzxcvbnmasdfghjklqwerty",
		"x",
		strings.Repeat("a", 120), // forces a version with version information
	} {
		m, err := Encode(payload)
		if err != nil {
			t.Fatalf("Encode(%d bytes): %v", len(payload), err)
		}
		size := m.Size()
		if (size-17)%4 != 0 || size < 21 || size > 57 {
			t.Fatalf("size %d is not 17+4v for a version in 1..10", size)
		}

		// The three finders, each a 7x7 ring with a 3x3 core.
		for _, origin := range [][2]int{{0, 0}, {size - 7, 0}, {0, size - 7}} {
			ox, oy := origin[0], origin[1]
			for dy := 0; dy < 7; dy++ {
				for dx := 0; dx < 7; dx++ {
					edge := dx == 0 || dx == 6 || dy == 0 || dy == 6
					core := dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4
					if want := edge || core; m[oy+dy][ox+dx] != want {
						t.Fatalf("finder at (%d,%d): module (%d,%d) is %v, want %v",
							ox, oy, dx, dy, m[oy+dy][ox+dx], want)
					}
				}
			}
		}

		// The timing lines alternate, starting dark at index 8 (which is even).
		for i := 8; i < size-8; i++ {
			if m[6][i] != (i%2 == 0) || m[i][6] != (i%2 == 0) {
				t.Fatalf("the timing pattern breaks at %d", i)
			}
		}

		// The module the standard fixes dark.
		if !m[size-8][8] {
			t.Error("the dark module at (8, size-8) is light")
		}

		// The separators: the ring immediately around each finder is light.
		for _, origin := range [][2]int{{0, 0}, {size - 7, 0}, {0, size - 7}} {
			ox, oy := origin[0], origin[1]
			for d := -1; d <= 7; d++ {
				for _, p := range [][2]int{{ox + d, oy - 1}, {ox + d, oy + 7}, {ox - 1, oy + d}, {ox + 7, oy + d}} {
					x, y := p[0], p[1]
					if x < 0 || y < 0 || x >= size || y >= size {
						continue
					}
					if m[y][x] {
						t.Fatalf("the separator around the finder at (%d,%d) is dark at (%d,%d)",
							ox, oy, x, y)
					}
				}
			}
		}
	}
}

// TestAlignmentPatternsSitOnTheirCentres.
func TestAlignmentPatternsSitOnTheirCentres(t *testing.T) {
	// 46 bytes needs version 5 at level M, which has one alignment pattern.
	m, err := Encode(strings.Repeat("b", 46))
	if err != nil {
		t.Fatal(err)
	}
	version := (m.Size() - 17) / 4
	for _, cy := range alignmentCentres[version] {
		for _, cx := range alignmentCentres[version] {
			// The three centres that fall on a finder are omitted by the standard.
			onFinder := (cx <= 8 && cy <= 8) ||
				(cx >= m.Size()-9 && cy <= 8) ||
				(cx <= 8 && cy >= m.Size()-9)
			if onFinder {
				continue
			}
			if !m[cy][cx] {
				t.Errorf("the alignment centre (%d,%d) is light", cx, cy)
			}
			for d := -2; d <= 2; d++ {
				if !m[cy-2][cx+d] || !m[cy+2][cx+d] || !m[cy+d][cx-2] || !m[cy+d][cx+2] {
					t.Errorf("the alignment ring at (%d,%d) is broken", cx, cy)
				}
			}
			if m[cy-1][cx] || m[cy+1][cx] || m[cy][cx-1] || m[cy][cx+1] {
				t.Errorf("the alignment gap at (%d,%d) is dark", cx, cy)
			}
		}
	}
}

// --- ④ round trip -------------------------------------------------------------------

// readMatrix is the test's own reader. It finds the mask from the format string
// rather than being told, un-masks, walks the zigzag, de-interleaves and parses
// the byte-mode header.
func readMatrix(t *testing.T, m Matrix) string {
	t.Helper()
	size := m.Size()
	version := (size - 17) / 4
	plan := plansM[version]

	// The format string, read from the copy along the top-left.
	raw := 0
	for i := 0; i < 15; i++ {
		var dark bool
		switch {
		case i < 6:
			dark = m[i][8]
		case i == 6:
			dark = m[7][8]
		case i == 7:
			dark = m[8][8]
		case i == 8:
			dark = m[8][7]
		default:
			dark = m[8][14-i]
		}
		if dark {
			raw |= 1 << uint(i)
		}
	}
	mask := -1
	for candidate := 0; candidate < 8; candidate++ {
		if formatInfo(candidate) == raw {
			mask = candidate
		}
	}
	if mask < 0 {
		t.Fatalf("the format string %015b is not a level-M format string", raw)
	}

	// Rebuild the reserved map by drawing the function patterns again — the
	// reader has to know which modules are data, and the standard says which.
	b := newBuilder(size)
	b.finders()
	b.timing()
	b.alignment(version)
	b.set(8, size-8, true)
	b.reserveFormat()
	if version >= 7 {
		b.versionBlocks(version)
	}

	bits := make([]bool, 0, plan.total*8)
	upward := true
	for right := size - 1; right >= 0; right -= 2 {
		if right == 6 {
			right--
		}
		for step := 0; step < size; step++ {
			y := step
			if upward {
				y = size - 1 - step
			}
			for dx := 0; dx < 2; dx++ {
				x := right - dx
				if b.reserved[y][x] {
					continue
				}
				dark := m[y][x]
				if maskAt(mask, y, x) {
					dark = !dark
				}
				bits = append(bits, dark)
			}
		}
		upward = !upward
	}

	codewords := make([]byte, plan.total)
	for i := 0; i < plan.total*8 && i < len(bits); i++ {
		if bits[i] {
			codewords[i/8] |= 1 << uint(7-i%8)
		}
	}

	// De-interleave back into blocks, keeping only the data half.
	sizes := make([]int, 0, plan.blocks())
	for i := 0; i < plan.g1Blocks; i++ {
		sizes = append(sizes, plan.g1Data)
	}
	for i := 0; i < plan.g2Blocks; i++ {
		sizes = append(sizes, plan.g2Data)
	}
	blocks := make([][]byte, len(sizes))
	longest := 0
	for i, n := range sizes {
		blocks[i] = make([]byte, 0, n)
		if n > longest {
			longest = n
		}
	}
	at := 0
	for i := 0; i < longest; i++ {
		for j, n := range sizes {
			if i < n {
				blocks[j] = append(blocks[j], codewords[at])
				at++
			}
		}
	}
	data := make([]byte, 0, plan.dataCodewords())
	for _, block := range blocks {
		data = append(data, block...)
	}

	// The byte-mode header, then the payload.
	read := func(offset, count int) int {
		value := 0
		for i := 0; i < count; i++ {
			bit := offset + i
			if data[bit/8]&(1<<uint(7-bit%8)) != 0 {
				value |= 1 << uint(count-1-i)
			}
		}
		return value
	}
	if mode := read(0, 4); mode != 0b0100 {
		t.Fatalf("the mode indicator is %04b, want byte mode 0100", mode)
	}
	cc := charCountBits(version)
	length := read(4, cc)
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		out[i] = byte(read(4+cc+8*i, 8))
	}
	return string(out)
}

// TestThePayloadComesBackOutOfTheSymbol.
func TestThePayloadComesBackOutOfTheSymbol(t *testing.T) {
	for _, payload := range []string{
		"x",
		"https://ntfy.sh/tossos-mnzxcvbnmasdfghjklqwerty",
		"https://ntfy.internal.example.com:8443/tossos-mnzxcvbnmasdfghjklqwerty",
		strings.Repeat("Z", 120),
		strings.Repeat("q", 213), // the largest a version-10 level-M symbol holds
	} {
		m, err := Encode(payload)
		if err != nil {
			t.Fatalf("Encode(%d bytes): %v", len(payload), err)
		}
		if got := readMatrix(t, m); got != payload {
			t.Errorf("round trip of %d bytes came back as %d bytes:\n got %q\nwant %q",
				len(payload), len(got), got, payload)
		}
	}
}

// TestTheSubscribeAddressFitsWithRoomToSpare.
//
// The address a075 generates is 46 bytes. This pins which version that lands on,
// so a change that quietly grew the channel — or the prefix — shows up as a
// bigger symbol rather than as a phone that will not focus.
func TestTheSubscribeAddressFitsWithRoomToSpare(t *testing.T) {
	m, err := Encode("https://ntfy.sh/tossos-mnzxcvbnmasdfghjklqwerty")
	if err != nil {
		t.Fatal(err)
	}
	if version := (m.Size() - 17) / 4; version > 4 {
		t.Errorf("the public subscribe address needs version %d (%dx%d modules); "+
			"it used to fit in 4 or less and a denser symbol is harder to scan",
			version, m.Size(), m.Size())
	}
}

// TestAPayloadTooLongIsRefusedRatherThanTruncated.
func TestAPayloadTooLongIsRefusedRatherThanTruncated(t *testing.T) {
	if _, err := Encode(strings.Repeat("z", 214)); err == nil {
		t.Fatal("a payload past version 10's capacity was accepted")
	}
}

// TestEveryVersionThisEncoderClaimsIsReachable.
//
// The plan table and the capacity arithmetic have to agree, and a version whose
// row is wrong shows up as a payload landing on a neighbour.
func TestEveryVersionThisEncoderClaimsIsReachable(t *testing.T) {
	seen := map[int]bool{}
	for n := 1; n <= 213; n++ {
		m, err := Encode(strings.Repeat("w", n))
		if err != nil {
			t.Fatalf("%d bytes: %v", n, err)
		}
		seen[(m.Size()-17)/4] = true
	}
	for version := 1; version <= MaxVersion; version++ {
		if !seen[version] {
			t.Errorf("no payload length lands on version %d; its table row may be wrong", version)
		}
	}
}

// TestTheChosenMaskIsTheLowestScoring.
func TestTheChosenMaskIsTheLowestScoring(t *testing.T) {
	payload := "https://ntfy.sh/tossos-mnzxcvbnmasdfghjklqwerty"
	m, err := Encode(payload)
	if err != nil {
		t.Fatal(err)
	}
	chosen := penalty(m)
	version, _ := smallestVersion(len(payload))
	plan := plansM[version]
	interleaved := interleave(encodeBits(payload, version, plan.dataCodewords()), plan)
	for mask := 0; mask < 8; mask++ {
		if p := penalty(place(version, interleaved, mask)); p < chosen {
			t.Errorf("mask %d scores %d, lower than the chosen %d", mask, p, chosen)
		}
	}
}

// TestTheDataModuleCountMatchesTheCapacityTable.
//
// This is the check that stands behind the round trip, because it does not share
// anything with it. The number of modules the placement walk considers free is
// derived from the drawing code — finders, separators, timing, alignment, format,
// version — and the number it MUST be is derived from the standard's codeword
// table: total codewords × 8, plus that version's remainder bits.
//
// Two independent sources for one number. A mis-drawn alignment pattern, a
// version block in the wrong place, or a format reservation that is one module
// short all move the first and not the second.
func TestTheDataModuleCountMatchesTheCapacityTable(t *testing.T) {
	for version := 1; version <= MaxVersion; version++ {
		size := 17 + 4*version
		b := newBuilder(size)
		b.finders()
		b.timing()
		b.alignment(version)
		b.set(8, size-8, true)
		b.reserveFormat()
		if version >= 7 {
			b.versionBlocks(version)
		}

		free := 0
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				if !b.reserved[y][x] {
					free++
				}
			}
		}
		want := plansM[version].total*8 + remainderBits[version]
		if free != want {
			t.Errorf("version %d: the walk sees %d free modules, the codeword table says %d "+
				"(%d codewords × 8 + %d remainder)", version, free, want,
				plansM[version].total, remainderBits[version])
		}
	}
}

// TestThePlacementStartsWhereTheStandardSaysItDoes.
//
// The zigzag begins at the bottom-right module and moves upward in two-column
// strips. A walk that started anywhere else would still round-trip through this
// package's own reader, so it is asserted directly.
func TestThePlacementStartsWhereTheStandardSaysItDoes(t *testing.T) {
	// One codeword whose first bit is 1 and whose rest are 0, in a version-1
	// symbol, with mask 0 applied by hand so the expected value is computable.
	version := 1
	size := 17 + 4*version
	codewords := make([]byte, plansM[version].total)
	codewords[0] = 0x80

	for mask := 0; mask < 8; mask++ {
		m := place(version, codewords, mask)
		want := true
		if maskAt(mask, size-1, size-1) {
			want = !want
		}
		if m[size-1][size-1] != want {
			t.Errorf("mask %d: the first data bit is not at the bottom-right module", mask)
		}
	}
}
