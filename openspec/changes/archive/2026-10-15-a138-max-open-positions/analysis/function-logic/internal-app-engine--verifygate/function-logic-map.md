# Function Logic Map: `verifyGate`

- Source: `internal/app/engine/interlock.go` (491-573)
- Qualified: `verifyGate`
- AST evidence: `ast.json`(현재) — 현재 AST branches 11 · returns 12
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 389371bc84e21e83…) — 편집 전 갈래 수 10
- Risk scan: `risk-pattern-report.md`

**역할.** 게이트 ON 기동 인터록의 단계 판정.

**a138 편집.** 7c `checkGuardianOpenPositionLimit` 한 단 — 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if facts.guardian == nil {` (if 493) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B2 | `if err := status.Limits.Validate(); err != nil {` (if 501) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B3 | `if err := checkTradingPolicy(facts.trading); err != nil {` (if 508) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B4 | `if err != nil {` (if 514) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B5 | `if accountRef == "" {` (if 526) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B6 | `if err := att.Verify(facts.now, accountRef, RequiredEndpoints()); err != nil {` (if 531) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B7 | `if err := checkGuardianLimits(facts.guardian, status.Limits); err != nil {` (if 537) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B8 | `if err := checkGuardianDailyBuyLimit(facts.guardian, *status); err != nil {` (if 541) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B9 | `if err := checkGuardianOpenPositionLimit(facts.guardian, *status); err != nil {` (if 545) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B10 | `if err := checkGateway(facts.gateway); err != nil {` (if 550) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |
| B11 | `if err := checkKeyCapableTransport(facts.transport); err != nil {` (if 553) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock|Gate|A138'` · `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` · `TestA138AnUnusableOpenPositionMaximumRefusesToStart` |

Exact AST return positions: 494 · 502 · 509 · 515 · 527 · 532 · 538 · 542 · 546 · 551 · 554 · 572

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 12종 — 대표: `RequiredEndpoints`, `att.Verify`, `attest.Load`, `checkGateway`, `checkGuardianDailyBuyLimit`, `checkGuardianLimits`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
