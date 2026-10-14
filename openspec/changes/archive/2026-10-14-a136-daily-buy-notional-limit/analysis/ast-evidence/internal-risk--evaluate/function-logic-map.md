# Function Logic Map: `Evaluate`

- Source: `internal/risk/chain.go` (131-144, base 7b073f55)
- Qualified: `Evaluate`
- 위치: `analysis/ast-evidence/` — **proposal 단계 근거(게이트 번들 아님)**. 편집하게 되는 함수는 구현 착수 때 `analysis/function-logic/` 에 정식 번들을 만든다.
- AST evidence: `ast.json` (`source_sha256` 43b2ef76b276cc56…) — AST branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`
- 용도: a136 proposal·design 의 주장 「매도는 진입 체인(따라서 새 일일 매수 한도 단계)에 닿지 않는다」의 근거. **a136 은 이 함수를 편집하지 않을 계획**(design D3: 새 검사는 체인 rung 이 아니라 발급 원자 트랜잭션에 둠).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.Intent.Side` | `BUY` · `SELL` (그 밖은 preflight 거부) | 호출자(발급자) | preflight 가 INPUT_UNAVAILABLE |
| `in.Policy` | `Validate()` 통과 | Guardian 생성 시 정책 | 진입 preflight 에서만 검사 |
| `entryChain` | 고정 12단계 | 패키지 변수 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if d := preflight(in); !d.Allowed` (132) | 없음 | `at(StepPreflight, d)` | `TestUnusableInputsRefuseRatherThanDefault` · `TestPreflightAndReductionReportDistinctSteps` |
| B2 | `if in.Intent.Side == SideSell` (135) | 없음 | `at(StepReduction, evaluateReduction(in))` — **매도는 entryChain 을 돌지 않음** | `TestReductionPassesEveryEntryOnlyGate` · `TestReductionIsNeverCostGated` |
| B3 | `range entryChain` (138) | 없음 | — | `TestChainOrderIsFixed` |
| B4 | `if d := s.eval(in); !d.Allowed` (139) | 없음 | `at(s.name, d)` 첫 실패 정지 | `TestFirstFailureStopsTheChain` |

Exact AST return positions: 133 · 136 · 140 · 143 (`allow()`)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `preflight` | 사용 불가 입력 거부 | 반환값만 | AST calls |
| `evaluateReduction` | 매도 = 보유 이하 reduce-only 만 판정 | 반환값만 | AST calls · chain.go |
| `s.eval` | 진입 12단계 | 반환값만 | AST calls |
| `at` | 정지 단계 도장 | — | AST calls |

## State mutations and fallbacks

- 순수 함수. 상태 변경 · fallback 없음.

## Safety conclusion

- Safe edit boundary: a136 은 편집하지 않음. 새 한도를 entryChain rung 으로 넣으면 B2 때문에 매도에 닿지 않는 것은 구조적으로 참이지만, 체인 입력은 「발급 시점에 한 번」 읽히므로 동시 발급 사이의 경쟁을 못 막음 → design D3 가 예약 트랜잭션을 택한 이유.
- High-risk impact: no (무편집)
