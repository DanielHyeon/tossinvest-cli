# Function Logic Map: `New`

- Source: `internal/execgw/gateway.go` (148-189)
- Revision: current — HEAD `648df8ef`; source_sha256 `9601d6562e363a2a5c70f69eccd02e24b5c8d5e216412b8fb9ea6d213ef36832`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 8 · returns 4 · calls 7
- Exact AST return positions: 151:3, 153:3, 155:3, 188:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | New | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | switch at 149:2 | `switch {` | evaluated 112/222 |
| B2 | case at 150:2 | `case opts.Journal == nil:` | NOT entered 0/222 |
| B3 | case at 152:2 | `case opts.Trading == nil:` | NOT entered 0/222 |
| B4 | case at 154:2 | `case strings.TrimSpace(opts.AccountRef) == "":` | NOT entered 0/222 |
| B5 | if at 158:2 | `if protectionCheckForTest == nil && !opts.forceReadinessAdapterForTest {` | entered 105/222 |
| B6 | if at 179:2 | `if g.clk == nil {` | NOT entered 0/222 |
| B7 | if at 182:2 | `if g.source == "" {` | NOT entered 0/222 |
| B8 | if at 185:2 | `if g.newID == nil {` | entered 112/222 |

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

- Safe edit boundary: New only.
- High-risk impact: reviewed and regression-tested.
