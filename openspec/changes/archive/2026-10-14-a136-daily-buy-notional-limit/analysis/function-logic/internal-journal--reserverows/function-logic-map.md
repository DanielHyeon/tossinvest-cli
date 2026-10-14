# Function Logic Map: `reserveRows`

- Source: `internal/journal/reservations.go` (323-375, 편집 뒤)
- Qualified: `reserveRows`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 933d37c95a6212fe…, 편집 전 분기 수 6) — 현재 branches 7 · returns 7
- Risk scan: `risk-pattern-report.md`

**역할.** 예약 tx 의 뒤 절반: 결정 재확인 → 종류별 총계(checkLimits) → 행 삽입 → 버전. 생산 호출자 5(모든 예약 tx — codegraph-baseline.md).

**a136 편집.** checkLimits 통과 뒤 `checkDailyBuyNotional(ctx, tx, req, plan, now)` 호출 한 문장 추가(5b). 거부면 아무것도 삽입하지 않고 반환 — 호출자 tx 롤백. **이 함수는 FLM 전에 편집됐다(freeze 리뷰 P1-3, review.md 에 기록)** — 이 번들은 그 뒤 base 트리에서 편집 전 AST 를 뽑아 사후 대조한 것. **분기 번호 대응(편집 전 → 뒤, 구현 리뷰 P3-6):** base B1·B2·B3 → B1·B2·B3 그대로, **B4 = a136 신규**(`checkDailyBuyNotional` 거부 갈래), base B4(range rows) → B5, base B5(INSERT 오류) → B6, base B6(버전 오류) → B7. 기존 분기의 조건 · 반환 무변경(리뷰가 ast-base.json 과 대조).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req` · `plan` | build 검증 통과 | 호출자 | — |
| `now` | journal 주입 시계 | j.clk | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := checkDecisionReservable(ctx, tx, plan, now); err != nil {` (if 326) | — | 분기 본문 참조 | `TestReservationRequiresARecordedUnexpiredDecision` |
| B2 | `if err != nil {` (if 332) | — | 분기 본문 참조 | `go test ./internal/journal -run TestA136` · 예약 시험군 |
| B3 | `if err := plan.checkLimits(held); err != nil {` (if 335) | — | 분기 본문 참조 | `TestReachingTheAggregateLimitIsRefused` |
| B4 | `if err := checkDailyBuyNotional(ctx, tx, req, plan, now); err != nil {` (if 339) | — | 분기 본문 참조 | `TestA136ReachingTheDailyBuyCeilingIsRefused` · `TestA136AnExposureRaisingReservationWithoutTheCeilingIsRefused` · `TestA136TheManualBuySnapshotCountsAndMustBeFresh` |
| B5 | `for _, row := range plan.rows {` (range 345) | — | 분기 본문 참조 | `TestReservationArithmeticIsExact` |
| B6 | `if _, err := tx.ExecContext(ctx,` (if 346) | INSERT risk_reservations | 분기 본문 참조 | 기존(DB 오류) |
| B7 | `if err != nil {` (if 370) | — | 분기 본문 참조 | 기존(DB 오류) |

Exact AST return positions: 327 · 333 · 336 · 340 · 354 · 371 · 374

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkDecisionReservable` | 결정 유효 | 오류 반환 | AST |
| `heldTotals` · `plan.checkLimits` | 종류별 총계 ≥ | ErrReservationLimitExceeded | AST |
| `checkDailyBuyNotional`(a136 새 leaf) | 하루 매수 총액 ≥ | ErrReservationLimitExceeded / ErrInvalidRequest / 오류 | daily_buy.go |

## State mutations and fallbacks

- INSERT 만(같은 tx). fallback 없음.

## Safety conclusion

- Safe edit boundary: 판정 호출 하나 추가 — 기존 분기 · 순서 무변경, 새 판정은 삽입 앞(거부 시 0행).
- High-risk impact: yes — 예약 권위. RED · 변이 원장 대상
