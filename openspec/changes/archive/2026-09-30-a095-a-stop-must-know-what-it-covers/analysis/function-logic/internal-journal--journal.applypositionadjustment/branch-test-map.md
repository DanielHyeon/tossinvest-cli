# Branch Test Map: `Journal.ApplyPositionAdjustment`

- Source: `internal/journal/position_adjustments.go`

> **「진입 실측」은 측정값이다** — 네 패키지의 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a095가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:204` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:210` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:219` `switch {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:220` `case err == nil:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:222` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:226` `case errors.Is(err, ErrAdjustmentNotFound):` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:227` `default:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:236` `switch {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:237` `case err == nil:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:239` `if current.State == PositionClosed {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:246` `case errors.Is(err, ErrPositionNotFound):` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:247` `default:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:251` `if same, cerr := sameDecimal(held, req.ExpectedPrevQuantity); cerr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:253` `} else if !same {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:253` `} else if !same {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:261` `if err := tx.QueryRowContext(ctx,` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:267` `if watermark != req.ExpectedFillWatermark {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:284` `if !adjustInPlace {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:286` `if current.ID != "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:318` `if result.OpenedInstance {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B21 | `:319` `if _, err := tx.ExecContext(ctx, `` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B22 | `:332` `if _, err := tx.ExecContext(ctx, `` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B23 | `:345` `if !result.OpenedInstance {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B24 | `:346` `if _, err := tx.ExecContext(ctx, `` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B25 | `:357` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B26 | `:363` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B27 | `:366` `if err := tx.Commit(); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 6개**: B2, B5, B7, B12, B25, B26
**자체 블록 없는 분기 5개**: B3, B8, B16, B21, B24 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
