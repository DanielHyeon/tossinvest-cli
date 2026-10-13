# Function Logic Map: `closeExitStateOnAdjustmentTx`

- Source: `internal/journal/position_adjustments.go` (`397`–`433`, 이 change 무편집)
- AST evidence: `ast.json` (`source_sha256` d4af875851e00c93…) — branches 6 · returns 7
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 4판(리뷰 4차 P2-6 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 대사 조정이 포지션을 0 으로 닫을 때 exit 상태를 완료로 표시하고 ADJUSTMENT_CLOSED 를 남김.
호출자: `ApplyPositionAdjustment`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:399` `if !isZeroDecimal(newQuantity) {` | — | — | 무편집 갈래 |
| B2 | `:408` `if errors.Is(err, sql.ErrNoRows) {` | — | — | 무편집 갈래 |
| B3 | `:413` `if err != nil {` | — | — | 무편집 갈래 |
| B4 | `:416` `if done != 0 {` | — | — | 무편집 갈래 |
| B5 | `:420` `if _, err := tx.ExecContext(ctx,` | — | — | 무편집 갈래 |
| B6 | `:425` `if err := appendExitEventTx(ctx, tx, exitEventRow{` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 근거 자리 참조.

## a130 편집 계획 (4판)

**인용 근거만 · 무편집.** 근거 자리: 이미 completed 면 돌아감(`done != 0`), 아니면 `completed = 1` UPDATE 뒤 `ExitEventAdjustmentClosed` — 그래서 같은 인스턴스에서 COMPLETED(ApplyExitFill B14 의 `!state.Completed`)와 배타적(design D4 모집단).

## Safety conclusion

- Safe edit boundary: 무편집.
- High-risk impact: no(무편집) — 인용 근거.
