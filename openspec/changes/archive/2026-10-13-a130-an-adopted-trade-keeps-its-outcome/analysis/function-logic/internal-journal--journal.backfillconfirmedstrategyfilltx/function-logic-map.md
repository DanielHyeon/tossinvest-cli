# Function Logic Map: `Journal.backfillConfirmedStrategyFillTx`

- Source: `internal/journal/strategy_dispatch_runtime.go` (편집 전 `1297`–`1385`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` 04b461130389895e…) — branches 15 · returns 14
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 확정된 전략 매수 체결을 원장에 되메움(BUY 만, :1303 · :1316).
호출자: `linkConfirmedStrategyDispatchTx`(:1095).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:1307` `if errors.Is(err, sql.ErrNoRows) {` | — | — | 무편집 갈래 |
| B2 | `:1310` `if err != nil {` | — | — | 무편집 갈래 |
| B3 | `:1313` `if failClosed != 0 {` | — | — | 무편집 갈래 |
| B4 | `:1320` `if fill.Terminal {` | — | — | 무편집 갈래 |
| B5 | `:1321` `if _, err := releaseReservationsForOrder(ctx, tx, a.brokerOrderID, a.intentID,` | — | — | 무편집 갈래 |
| B6 | `:1327` `if err := j.applyRiskBucketFillInTx(ctx, tx, fill); err != nil {` | — | — | 무편집 갈래 |
| B7 | `:1330` `if err := j.releaseTerminalRiskBucketOrderInTx(ctx, tx, fill); err != nil {` | — | — | 무편집 갈래 |
| B8 | `:1338` `if allHooksMissing {` | — | — | 무편집 갈래 |
| B9 | `:1344` `if orderErr != nil {` | — | — | 무편집 갈래 |
| B10 | `:1348` `if err := j.runApplyHooks(ctx, tx, fill); err != nil {` | — | — | 무편집 갈래 |
| B11 | `:1358` `if err != nil {` | — | — | 무편집 갈래 |
| B12 | `:1361` `if err := requireStrategyDispatchExactRows(rebased, 1, "strategy fill ownership rebase"); ` | — | — | 무편집 갈래 |
| B13 | `:1364` `if _, err := tx.ExecContext(ctx, `UPDATE fill_snapshots SET committed_at=? WHERE order_id=` | — | — | 무편집 갈래 |
| B14 | `:1374` `if campaignBound {` | — | — | 무편집 갈래 |
| B15 | `:1379` `if err := ApplyPositionCampaignFill(ctx, handle, fill); err != nil {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

**인용 근거만 · 무편집.** 근거: `Side: "BUY"` 체결만 적용(:1303 · :1316) — 매수는 포지션을 닫지 못하므로 B10(:1348) `runApplyHooks` 에서 성과 동결 거절이 생기지 않음. ctx 에 수집 목록을 싣지 않음.

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: no(무편집) — 인용 근거.
