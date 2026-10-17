# Function Logic Map: `Journal.ApplyPositionAdjustment`

- Source: `internal/journal/position_adjustments.go` (편집 전 `202`–`372`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` d4af875851e00c93…) — branches 27 · returns 17
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** compare-and-append 대사 조정 — 프로젝션을 계좌 값으로 맞추고 조정 행을 덧붙임.
호출자: 생산 `Converger.ConvergeQuantities`(converge.go) · 외부 fold-in(external.go) 외.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:204` `if err != nil {` | — | — | 무편집 갈래 |
| B2 | `:210` `if err != nil {` | — | — | 무편집 갈래 |
| B3 | `:219` `switch {` | — | — | 무편집 갈래 |
| B4 | `:220` `case err == nil:` | — | — | 무편집 갈래 |
| B5 | `:222` `if err != nil {` | — | — | 무편집 갈래 |
| B6 | `:226` `case errors.Is(err, ErrAdjustmentNotFound):` | — | — | 무편집 갈래 |
| B7 | `:227` `default:` | — | — | 무편집 갈래 |
| B8 | `:236` `switch {` | — | — | 무편집 갈래 |
| B9 | `:237` `case err == nil:` | — | — | 무편집 갈래 |
| B10 | `:239` `if current.State == PositionClosed {` | — | — | 무편집 갈래 |
| B11 | `:246` `case errors.Is(err, ErrPositionNotFound):` | — | — | 무편집 갈래 |
| B12 | `:247` `default:` | — | — | 무편집 갈래 |
| B13 | `:251` `if same, cerr := sameDecimal(held, req.ExpectedPrevQuantity); cerr != nil {` | — | — | 무편집 갈래 |
| B14 | `:253` `} else if !same {` | — | — | 무편집 갈래 |
| B15 | `:253` `} else if !same {` | — | — | 무편집 갈래 |
| B16 | `:261` `if err := tx.QueryRowContext(ctx,` | — | — | 무편집 갈래 |
| B17 | `:267` `if watermark != req.ExpectedFillWatermark {` | — | — | 무편집 갈래 |
| B18 | `:284` `if !adjustInPlace {` | — | — | 무편집 갈래 |
| B19 | `:286` `if current.ID != "" {` | — | — | 무편집 갈래 |
| B20 | `:318` `if result.OpenedInstance {` | — | — | 무편집 갈래 |
| B21 | `:319` `if _, err := tx.ExecContext(ctx, `` | — | — | 무편집 갈래 |
| B22 | `:332` `if _, err := tx.ExecContext(ctx, `` | — | — | 무편집 갈래 |
| B23 | `:345` `if !result.OpenedInstance {` | — | — | 무편집 갈래 |
| B24 | `:346` `if _, err := tx.ExecContext(ctx, `` | — | — | 무편집 갈래 |
| B25 | `:357` `if err != nil {` | — | — | 무편집 갈래 |
| B26 | `:363` `if err != nil {` | — | — | 무편집 갈래 |
| B27 | `:366` `if err := tx.Commit(); err != nil {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

**인용 근거만 · 무편집.** a130 이 근거로 쓰는 자리: `:312` `NewAvgPrice: firstNonEmpty(req.NewAvgPrice, target.AvgPrice)` — 요청이 "" 이면 프로젝션의 기존 평균가를 이어 적음(늘어난 수량의 매입가 아님, design 사실절 · I-1). 조정 행의 `created_at` = 저널 시계(`j.nowString()`).

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: no(무편집) — 인용 근거.
