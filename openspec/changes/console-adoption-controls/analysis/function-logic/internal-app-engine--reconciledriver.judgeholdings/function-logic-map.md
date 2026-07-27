# Function Logic Map: `ReconcileDriver.judgeHoldings`

- Source: `internal/app/engine/adoption.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

High-risk(reconciliation). findings 구간만 수정: exclude 판정을 먼저 두고(스펙: exclude 항상 우선), 후보 조건을 (enabled ∨ Included(sym))로 확장. already-managed→RECONCILE→stale 게이트 순서 불변 — include가 신선·Stabiliser·Verified를 우회할 수 없는 구조적 근거. 빈 include는 기존 경로와 동일(§0.2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B2 | range | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B3 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B4 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B5 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B6 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B7 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B8 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B9 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B10 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B11 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B12 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B13 | range | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B14 | if | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |
| B15 | range | 없음 | — | TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(RED: adopted=0)·TestIncludeDoesNotBypassExclusion(RED: 사유 미표기)·기존 reconcileloop 16종 green·-race 629 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음(판정·분류만; 편입 tx는 adoptOne 소유 — 무변경)

## Safety conclusion

- Safe edit boundary: findings 분기 2개 재배열·확장만 — 수집·fold·adopt·알림 latch 무접촉
- High-risk impact: yes — reconciliation/audit 경로, Pre-Edit 선언·적대적 리뷰 하에 수정
