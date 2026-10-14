# Function Logic Map: `evaluateChainWithDailyLoss`

- Source: `internal/execgw/a137_daily_loss.go` (183-189, 편집 뒤)
- Qualified: `evaluateChainWithDailyLoss`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` d0869728c1f8f72d…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 5ebfb6c220c7568b…, 편집 전 분기수 1) — 현재 branches 1 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 체인 1회 + 측정 실패 상세 교체.

**a137 편집.** 구현 리뷰 수리 커밋에서 주석만 바뀜(본문 · 분기 불변, 1 → 1) — 표지 부착은 새 함수 `dailyLossChainRefusal` 가 맡음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in · measureErr | 체인 입력 · 측정 실패 | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !verdict.Allowed && measureErr != nil && verdict.Step == "daily_loss" && verdict.Reason == risk.ReasonInputUnavailable {` (if 185) | — | 분기 본문 참조 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller`(상세 교체) · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure`(앞 단 사유 보존) |

Exact AST return positions: 188

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluateChain` | 체인 | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 상세 문구.
- High-risk impact: yes(진입 판정 문구) — 판정 불변.
