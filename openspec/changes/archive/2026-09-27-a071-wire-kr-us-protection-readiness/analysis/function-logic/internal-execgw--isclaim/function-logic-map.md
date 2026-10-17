# Function Logic Map: `isClaim`

- Source: `internal/execgw/protection_test.go` (214-240)
- Revision: base `775c37cb` (HEAD 에 함수 없음); source_sha256 `7a9c1570ece61fdeac78438e9e5b49c64f27d987fe592083103e5aab85e5ab69`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 9 · returns 8 · calls 15
- Exact AST return positions: 218:3, 220:3, 222:3, 226:3, 233:3, 235:3, 237:3, 239:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | isClaim | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | switch at 216:2 | `switch {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B2 | case at 217:2 | `case strings.HasPrefix(trimmed, name+" "):` | base 소스의 분기 — HEAD 에 함수 없음 |
| B3 | case at 219:2 | `case strings.HasPrefix(trimmed, "var "+name+" "), strings.HasPrefix(trimmed, "var "+name+" ="):` | base 소스의 분기 — HEAD 에 함수 없음 |
| B4 | case at 221:2 | `case strings.HasPrefix(trimmed, "var wiredForTest ="), strings.HasPrefix(trimmed, "var WiredProtectionForTe...` | base 소스의 분기 — HEAD 에 함수 없음 |
| B5 | if at 225:2 | `if at < 0 {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B6 | switch at 231:2 | `switch {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B7 | case at 232:2 | `case strings.HasSuffix(before, "=="), strings.HasSuffix(before, "!="):` | base 소스의 분기 — HEAD 에 함수 없음 |
| B8 | case at 234:2 | `case strings.HasSuffix(before, "case"), strings.HasSuffix(before, "return"):` | base 소스의 분기 — HEAD 에 함수 없음 |
| B9 | case at 236:2 | `case strings.HasSuffix(before, "="), strings.HasSuffix(before, ":="), strings.HasSuffix(before, ":"):` | base 소스의 분기 — HEAD 에 함수 없음 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- Base-revision evidence records the removed or renamed scalar-test path.

## Safety conclusion

- Safe edit boundary: isClaim only.
- High-risk impact: reviewed and regression-tested.
