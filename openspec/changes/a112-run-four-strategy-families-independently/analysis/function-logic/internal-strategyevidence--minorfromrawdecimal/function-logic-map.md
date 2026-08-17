# Function Logic Map: `minorFromRawDecimal`

- Source: `internal/strategyevidence/breakout_bar.go`
- Source SHA-256: `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `minorFromRawDecimal(raw string, scale uint64) (uint64, error)`
- Source range: `822:1`–`861:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 10, returns 7, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Callers: `recomputeMinors` (breakout_bar.go:412, constructor write path) and `checkRawMinor` (breakout_bar.go:864, decoder read/replay path). Converts the broker's raw decimal string to integer minor units at the declared scale using only `×10 + digit` on `uint64` (`appendDecimalDigit`); no float, no rounding.
- Accepted grammar: `<digits>` or `<digits>.<digits>`; no sign, exponent, space or comma; integer part non-empty and without a redundant leading zero (`0` alone allowed); fraction shorter than `scale` is zero-padded, longer is refused (fail closed, amended decision 9).
- Any error text is wrapped by the caller into a `payloadFieldError` naming the raw field.

## Branches and early returns

Exact AST return nodes: `824, 832, 836, 839, 845, 856, 860` (6 refusals + the value).

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 823:2 | empty raw → refuse | `TestClosedBarRejectsSignedExponentOrPaddedRawDecimal` (`""` case) |
| B2 | if | 827:2 | raw contains `.` → split integer/fraction | every fractional fixture (`231.65`, `231.1`, `231.4350`) — `TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars`, `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` |
| B3 | if | 830:2 | empty integer part, non-digit characters, or trailing `.` → refuse | `TestClosedBarRejectsSignedExponentOrPaddedRawDecimal` (`+231.65`, `-231.65`, `2.3165e2`, ` 231.65`, `231.65 `, `231,65`, `231.`, `.65`, `23I.65`) |
| B4 | if | 835:2 | integer part longer than 1 with leading `0` → refuse | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` (`00231.4350`, `01`, `00`), `TestClosedBarRefusesLeadingZeroRawDecimal` |
| B5 | if | 838:2 | more fraction digits than `scale` → refuse | `TestClosedBarRejectsOverPreciseRawForTheDeclaredScale` (5 dp USD, 1 dp KRW), `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` (long over-precise raw refused by this rule first) |
| B6 | for | 842:2 | fold integer digits | every accepted raw |
| B7 | if | 844:3 | integer fold overflows `uint64` → refuse | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` (`18446744073709551616`@0/@4, 28-digit integer part of the 33-char raw refused by "does not fit in 64 bits") |
| B8 | for | 849:2 | append `scale` fraction positions (zero-padded) | every accepted raw with scale 4; scale 0 (KR, volume) skips the loop — `TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars` |
| B9 | if | 851:3 | position inside the given fraction → use its digit, else 0 | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` (`0.5`@4 → 5000, `0.0001`@4 → 1, `231.1`@4 → 2311000), `TestNewClosedBar1mEnvelopeDerivesScaleMinorsAndIdentityTogether` (4/4/1/2 dp raws) |
| B10 | if | 855:3 | scaling overflows `uint64` → refuse | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` (`1844674407370955.1616`@4 refused; `.1615` = 2^64−1 accepted) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `strings.IndexByte(raw, '.')` | 827, 831 | split point; second call re-checks the trailing-`.` case |
| `onlyDigits(integerPart)`, `onlyDigits(fractionPart)` | 830 | ASCII digit predicate (same file) |
| `appendDecimalDigit(value, digit)` | 843, 854 | overflow-guarded `value*10+digit`; both guards probed by the reviewers (`2^64−1` accepted, `+1` refused) |
| `errors.New`, `formatUint`, `len`, `uint64` | refusals | text only |

## State mutations and fallbacks

- None. Locals only (`integerPart`, `fractionPart`, `value`, `next`, `digit`, loop indices — AST 14 assignments, all local). No fallback: no rounding, no truncation, no default scale.

## Safety conclusion

- This function is the only conversion from broker decimal text to stored integer minor units, on both the write (constructor) and read (decoder) paths, so a bug here would corrupt price evidence silently. All ten branches are pinned by direct unit tests plus decoder-level tests; the leading-zero rule (B4) is a fail-closed contract carried into L1b's receipt check (review.md g3), and the former 32-char length bound was dropped as unreachable defence (g5) with the test asserting which rule now refuses long raws.
