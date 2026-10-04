# Function Logic Map: `canonicalIntegerValue`

- Source: `internal/strategyevidence/breakout_bar.go`
- Source SHA-256: `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `canonicalIntegerValue(text string) (uint64, error)`
- Source range: `788:1`–`818:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 8, returns 7, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Called from `payloadReader.integer` (breakout_bar.go:727) with the text of a `json.Number` taken from the canonical object. Because `decodeCanonicalObject` runs `canonicalJSON` first, the text is always in canonical form: `[-]digits` optionally followed by `e<int>` (e.g. `60000` → `6e4`, `1.50` → `15e-1`, `1.2345e4` → `12345`); it never contains `.` or a non-integer exponent (model.go:296–396).
- Contract (D1, accepted): integer-ness is judged by value, not spelling — `1.9080e4` is `19080` and accepted; fractions (`15e-1`), signs (`-1`) and values above 2^64−1 are refused.

## Branches and early returns

Exact AST return nodes: `794, 797, 800, 805, 809, 813, 817`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 790:2 | text carries an `e`/`E` exponent → split mantissa/exponent | `TestCanonicalIntegerValueBoundaries` (`6e4`, `1e19`, `2e19`, `15e-1`, `1e21`), `TestClosedBarIntegerRuleIsAboutTheValueNotTheSpelling` |
| B2 | if | 793:3 | exponent not parseable as int → refuse | not-applicable: unreachable on canonical input — `canonicalJSON` emits only integer exponents (defensive); no direct-call fixture |
| B3 | if | 796:3 | negative exponent → refuse (not a whole number) | `TestCanonicalIntegerValueBoundaries` (`15e-1`), `TestClosedBarRejectsDecimalPriceNumber` (`231.65` → `23165e-2`) |
| B4 | if | 799:3 | exponent > 20 → refuse (cannot fit 64 bits) | `TestCanonicalIntegerValueBoundaries` (`1e21`) |
| B5 | if | 804:2 | mantissa contains `.` → refuse | not-applicable: unreachable on canonical input — `canonicalJSON` never emits `.` (defensive); no direct-call fixture |
| B6 | if | 808:2 | `ParseUint` fails (sign, empty, non-digit, > 2^64−1) → refuse | `TestCanonicalIntegerValueBoundaries` (`-1`, `""`, `18446744073709551616`), `TestClosedBarRejectsDecimalPriceNumber` (negative minor) |
| B7 | for | 811:2 | multiply by 10 per exponent step | `TestCanonicalIntegerValueBoundaries` (`6e4`, `1e19` accepted) |
| B8 | if | 812:3 | next ×10 would overflow → refuse | `TestCanonicalIntegerValueBoundaries` (`2e19` refused, `1e19` accepted at the boundary; envelope-level `volume` `2e19` refused) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strings.IndexAny` | 790:14 |
| `strconv.Atoi` | 792:18 |
| `errors.New` | 794:14 |
| `errors.New` | 797:14 |
| `errors.New` | 800:14 |
| `strings.ContainsRune` | 804:5 |
| `errors.New` | 805:13 |
| `strconv.ParseUint` | 807:16 |
| `errors.New` | 809:13 |
| `errors.New` | 813:14 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `strings.IndexAny(text, "eE")` | 790 | exponent split |
| `strconv.Atoi(text[index+1:])` | 792 | exponent parse (integer by construction of canonical form) |
| `strings.ContainsRune(digits, '.')` | 804 | defensive fraction guard |
| `strconv.ParseUint(digits, 10, 64)` | 807 | mantissa parse; refuses sign/empty/overflow |
| `errors.New` | refusals | text only |

## State mutations and fallbacks

- None. Locals only (`digits`, `exponent`, `parsed`, `value`, `step` — AST 9 assignments, all local). No fallback: no truncation, no saturation, no float conversion.

## Safety conclusion

- Every integer field of both breakout payloads (minors, volume, clocks, revision, scale, interval) passes through this function on write, replay and read; six of eight branches are pinned by direct and envelope-level tests including the 2^64 boundary, and the two defensive branches (B2, B5) are unreachable because the input is canonical by construction. Both reviewers probed the bounds (`1e19` ok, `2e19`/`2^64`/`1e1000000` refused).
