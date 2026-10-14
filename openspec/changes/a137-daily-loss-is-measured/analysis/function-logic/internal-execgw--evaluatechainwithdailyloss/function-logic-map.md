# Function Logic Map: `evaluateChainWithDailyLoss`

- Source: `internal/execgw/a137_daily_loss.go` (177-183, base 리비전)
- Qualified: `evaluateChainWithDailyLoss`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 5ebfb6c220c7568b…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 5ebfb6c220c7568b…, 편집 전 분기수 1) — 현재 branches 1 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 체인 1회 + 측정 실패 상세 교체.

**a137 편집.** 본문 무편집 — c83241ff 의 diff hunk 가 함수 끝 줄(옛 쪽)에 닿아(바로 뒤에 새 함수 · 새 시험을 덧붙임) 게이트가 base 리비전 증거를 요구함. `ast.json` 은 `revision: base`(63fd0c92 blob, 분기 불변).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in · measureErr | 체인 입력 · 측정 실패 | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `return amount, true` (if 179) | — | 분기 본문 참조 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller`(상세 교체) · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure`(앞 단 사유 보존) |

Exact AST return positions: 182

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluateChain` | 체인 | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 상세 문구.
- High-risk impact: yes(진입 판정 문구) — 판정 불변.
