# Function Logic Map: `evaluateChainWithDailyLoss`

- Source: `internal/execgw/a137_daily_loss.go` (187-199, 편집 뒤)
- Qualified: `evaluateChainWithDailyLoss`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` d49003643c5f13fb…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 5ebfb6c220c7568b…, 편집 전 분기수 1) — 현재 branches 2 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 체인 1회 + 측정 실패 상세 교체 + 원인 판정.

**a137 편집.** 재리뷰 P2-N1: 둘째 반환값(손실 입력이 원인일 때만 측정 오류). 손실만 0(기준 통화)으로 바꿔 체인을 다시 돌려 여전히 daily_loss 입력 불가면(B2 추가) 원인은 자본 등 — 상세 · 표지 없음. 분기 1 → 2.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in · measureErr | 체인 입력 · 측정 실패 | 두 진입 문 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if verdict.Allowed \|\| measureErr == nil \|\| verdict.Step != "daily_loss" \|\| verdict.Reason != risk.ReasonInputUnavailable {` (if 189) | — | 분기 본문 참조 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller` · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure` |
| B2 | `if again := evaluateChain(probe); !again.Allowed && again.Step == "daily_loss" && again.Reason == risk.ReasonInputUnavailable {` (if 194) | — | 분기 본문 참조 | `TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked` |

Exact AST return positions: 190 · 195 · 198

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluateChain` | 체인(최대 2회) | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 상세 · 원인 판정.
- High-risk impact: yes(진입 분류) — 무효 자본은 오늘처럼 결함.
