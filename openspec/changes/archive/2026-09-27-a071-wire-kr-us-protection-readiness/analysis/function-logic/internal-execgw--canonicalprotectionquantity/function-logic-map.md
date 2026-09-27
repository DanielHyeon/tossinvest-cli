# Function Logic Map: `canonicalProtectionQuantity`

- Source: `internal/execgw/protection.go` (112-123)
- Revision: current — HEAD `648df8ef`; source_sha256 `71e4923e1301555808b3c65b437d1d20906f9d633d8eef52ac676a1433cd8267`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 2 · returns 3 · calls 5
- Exact AST return positions: 116:3, 120:3, 122:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| order intent quantity | positive canonical integral float64 up to 2^53-1 | signed order intent | return `(0,false)` before provider/broker |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 115:2 | `if canonical == "" \|\| strings.ContainsAny(canonical, ".eE-+") {` | entered 2/222 |
| B2 | if at 119:2 | `if err != nil \|\| quantity == 0 \|\| quantity > maximumExactlyRepresentableInteger \|\| float64(quantity) ...` | entered 1/222 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | decimal text empty or signed/fractional/exponent form | none | reject | canonical quantity unit matrix |
| N2 | parse fails, zero, above safe integer, or round-trip differs | none | reject | canonical quantity unit matrix |
| N3 | exact positive safe integer | none | uint64 plus true | canonical quantity unit matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decimalString`, `strconv.ParseUint` | canonicalize without float rounding authority | any failure rejects | CodeGraph + AST |

## State mutations and fallbacks

- Pure conversion; never calls readiness provider or broker.

## Safety conclusion

- Safe edit boundary: exact integral quantity validation before readiness
- High-risk impact: yes; fail closed before transport
