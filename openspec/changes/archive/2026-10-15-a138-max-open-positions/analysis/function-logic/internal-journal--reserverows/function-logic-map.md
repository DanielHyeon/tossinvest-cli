# Function Logic Map: `reserveRows`

- Source: `internal/journal/reservations.go` (328-384)
- Qualified: `reserveRows`
- AST evidence: `ast.json`(현재) — 현재 AST branches 8 · returns 8
- 편집 전 기준: `ast-base.json`(base, `source_sha256` f94ae7a5298c0c06…) — 편집 전 갈래 수 7
- Risk scan: `risk-pattern-report.md`

**역할.** 모든 예약 tx 가 지나는 한 함수 — 결정 확인 · 총계 · 하루 매수 · 삽입.

**a138 편집.** 5b(하루 매수) 다음에 5c `checkOpenPositions` 호출 한 줄 — 새 분기 하나(거부 시 반환).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := checkDecisionReservable(ctx, tx, plan, now); err != nil {` (if 331) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B2 | `if err != nil {` (if 337) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B3 | `if err := plan.checkLimits(held); err != nil {` (if 340) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B4 | `if err := checkDailyBuyNotional(ctx, tx, req, plan, now); err != nil {` (if 344) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B5 | `if err := checkOpenPositions(ctx, tx, req, plan); err != nil {` (if 348) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B6 | `for _, row := range plan.rows {` (range 354) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B7 | `if _, err := tx.ExecContext(ctx,` (if 355) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B8 | `if err != nil {` (if 379) | — | 분기 본문 참조 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |

Exact AST return positions: 332 · 338 · 341 · 345 · 349 · 363 · 380 · 383

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 14종 — 대표: `append`, `checkDailyBuyNotional`, `checkDecisionReservable`, `checkOpenPositions`, `fmt.Errorf`, `formatJournalTime`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
