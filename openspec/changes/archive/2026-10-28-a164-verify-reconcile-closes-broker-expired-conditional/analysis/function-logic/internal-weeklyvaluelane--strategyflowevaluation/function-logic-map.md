# Function Logic Map: `strategyflowEvaluation`

- Source: `internal/weeklyvaluelane/strategyflow_testseam.go` (147-150)
- AST evidence: `ast.json`(revision current) — AST branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`(규칙 일치 0)
- 작성 시점: a164 종결(2026-10-28, 종결 에이전트). 이 함수를 고친 것은 형제 착지 `3f8e34632`(feat(offline), STORY-TOS-a162)이고 a164 base `4cd7af19d` 보다 뒤라 gate ⑤ 창(base→워킹트리)에 들어옴. 어느 change 에도 번들이 없어(활성·archive 전수 검색 0) 새로 작성함. 구현 뒤 사후 작성한 증거이며 구현 판단에 쓰이지 않음.

**역할.** 시험 seam 내부 헬퍼. 3f8e34632 에서 `ordinal` 인자가 생기고, 계획·예약·평가를 직접 하던 본문이 `strategyflowEvaluationWithPlan(…, ordinal, currency, currency, nil)` 위임으로 바뀜. 시장별 통화(KR→KRW, US→USD)를 계좌·호가 통화 둘 다로 주고 FX 는 nil(같은 통화).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `evidence.Market` | `MarketKR`·`MarketUS` | 호출자 fixture | 다른 값이면 통화가 빈 문자열 — 위임 대상의 계획 생성이 거절(이 함수는 판정 안 함) |
| `ordinal` | 1~7 | 호출자 | 범위 밖은 위임 대상이 `RefusalPlanExhausted` 로 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 분기 0 — 위임 한 줄 | — | `strategyflowEvaluationWithPlan` 결과 그대로 | `TestStrategyflowFollowOnFixtureBuildsNativePrefixThroughOrdinalSeven`(ordinal 3~7, KR·US) · `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures/{KR,US}-weekly`(ordinal 1) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyflowEvaluationWithPlan` | 시장 통화로 계획 포함 요청 생성 | 오류를 감싸지 않고 그대로 반환 | AST calls 149 |

## State mutations and fallbacks

- 상태 변경 0. 전역·디스크·네트워크 접근 없음(위임 대상이 메모리 안에서 요청 값을 만듦).

## Safety conclusion

- Safe edit boundary: `//go:build tossos_testseams` 파일의 시험 fixture — 생산 빌드에 들어가지 않음.
- High-risk impact: no — 주문·손절·사이징·인증 경로 무접촉(시험 seam 빌드 전용).
