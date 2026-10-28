# Function Logic Map: `StrategyflowUSFixture`

- Source: `internal/weeklyvaluelane/strategyflow_testseam.go` (45-47)
- AST evidence: `ast.json`(revision current) — AST branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`(규칙 일치 0)
- 작성 시점: a164 종결(2026-10-28, 종결 에이전트). 이 함수를 고친 것은 형제 착지 `3f8e34632`(feat(offline), STORY-TOS-a162)이고 a164 base `4cd7af19d` 보다 뒤라 gate ⑤ 창(base→워킹트리)에 들어옴. 어느 change 에도 번들이 없어(활성·archive 전수 검색 0) 새로 작성함. 구현 뒤 사후 작성한 증거이며 구현 판단에 쓰이지 않음.

**역할.** 시험 seam 공개 fixture. 3f8e34632 에서 본문이 `strategyflowUSFixture(candidateID, evidenceDigest, 1)` 위임 한 줄로 바뀜 — 첫 회차(ordinal 1) US 주간 요청을 만드는 기존 동작을 서수 인자를 받는 내부 함수로 옮김.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `candidateID`, `evidenceDigest` | 시험이 주는 임의 문자열 | 호출 시험 | 검증은 위임 대상이 함 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 분기 0 — 위임 한 줄 | — | `strategyflowUSFixture` 결과 그대로 | `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures/US-weekly`(`internal/strategyflow`) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyflowUSFixture` | ordinal 1 로 US 주간 요청 생성 | 오류를 감싸지 않고 그대로 반환 | AST calls 46 |

## State mutations and fallbacks

- 상태 변경 0. 전역·디스크·네트워크 접근 없음(위임 대상이 메모리 안에서 요청 값을 만듦).

## Safety conclusion

- Safe edit boundary: `//go:build tossos_testseams` 파일의 시험 fixture — 생산 빌드에 들어가지 않음.
- High-risk impact: no — 주문·손절·사이징·인증 경로 무접촉(시험 seam 빌드 전용).
