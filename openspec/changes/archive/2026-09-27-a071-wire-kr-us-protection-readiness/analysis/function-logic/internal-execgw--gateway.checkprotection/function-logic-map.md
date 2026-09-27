# Function Logic Map: `Gateway.checkProtection`

- Source: `internal/execgw/protection.go` (89-110)
- Revision: current — HEAD `648df8ef`; source_sha256 `71e4923e1301555808b3c65b437d1d20906f9d633d8eef52ac676a1433cd8267`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 5 · returns 6 · calls 8
- Exact AST return positions: 91:3, 94:3, 97:3, 101:3, 107:3, 109:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | Gateway.checkProtection | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 90:2 | `if !plan.raisesExposure {` | entered 22/222 |
| B2 | if at 93:2 | `if g.protectionCheckForTest != nil {` | entered 84/222 |
| B3 | if at 96:2 | `if g.protectionReadiness == nil {` | entered 1/222 |
| B4 | if at 100:2 | `if !ok {` | entered 1/222 |
| B5 | if at 106:2 | `if refusal != nil {` | entered 1/222 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- No authority broadening; current behavior is covered by focused tests.

## Safety conclusion

- Safe edit boundary: Gateway.checkProtection only.
- High-risk impact: reviewed and regression-tested.
