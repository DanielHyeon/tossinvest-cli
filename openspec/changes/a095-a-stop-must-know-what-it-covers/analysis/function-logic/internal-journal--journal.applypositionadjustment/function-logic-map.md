# Function Logic Map: `Journal.ApplyPositionAdjustment`

- Source: `internal/journal/position_adjustments.go` (`202`–`372`)
- Qualified: `Journal.ApplyPositionAdjustment`
- AST evidence: `ast.json` (`source_sha256` d4af875851e00c93…)
- Risk scan: `risk-pattern-report.md`
- 분기 27 · return 17 · 호출 37

**역할.** 계좌 스냅샷이 알려 준 수량 변화를 원장에 반영한다. **앱에서 넣은 매수가 엔진에 들어오는 문이다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req.NewQuantity` | 브로커가 말한 수량 | 계좌 스냅샷 | `positions.quantity`를 덮는다 |
| `req.NewAvgPrice` | 브로커가 말한 평단 | 계좌 스냅샷 | **비면 `target.AvgPrice`를 그대로 이어붙인다**(`:312` `firstNonEmpty`) |
| `req.ExpectedPrevQuantity` | 낙관적 잠금 | 호출자 | 어긋나면 `ErrAdjustmentStale` |
| `watermark` | 체결 워터마크 | 원장 | 움직였으면 폐기 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:204` `if err != nil {` | `j.db.BeginTx`, `j.nowString` | :205 | 예 |
| B2 | if | `:210` `if err != nil {` | `fmt.Errorf`, `scanAdjustment`, `tx.QueryRowContext`, `tx.Rollback` | :211 | 아니오 |
| B3 | switch | `:219` `switch {` | — | — | — |
| B4 | case | `:220` `case err == nil:` | `scanPosition`, `tx.QueryRowContext` | — | 예 |
| B5 | if | `:222` `if err != nil {` | — | :223, :225 | 아니오 |
| B6 | case | `:226` `case errors.Is(err, ErrAdjustmentNotFound):` | `errors.Is` | — | 예 |
| B7 | case | `:227` `default:` | `scanPosition`, `tx.QueryRowContext` | :228 | 아니오 |
| B8 | switch | `:236` `switch {` | — | — | — |
| B9 | case | `:237` `case err == nil:` | — | — | 예 |
| B10 | if | `:239` `if current.State == PositionClosed {` | — | — | 예 |
| B11 | case | `:246` `case errors.Is(err, ErrPositionNotFound):` | `errors.Is` | — | 예 |
| B12 | case | `:247` `default:` | — | :248 | 아니오 |
| B13 | if | `:251` `if same, cerr := sameDecimal(held, req.ExpectedPrevQuantity); cerr != nil {` | `sameDecimal` | :252 | 예 |
| B14 | else | `:253` `} else if !same {` | — | — | 예 |
| B15 | if | `:253` `} else if !same {` | — | :254 | 예 |
| B16 | if | `:261` `if err := tx.QueryRowContext(ctx,` | `Scan`, `fmt.Errorf`, `tx.QueryRowContext` | :264 | — |
| B17 | if | `:267` `if watermark != req.ExpectedFillWatermark {` | `fmt.Sprintf` | :268 | 예 |
| B18 | if | `:284` `if !adjustInPlace {` | `int64` | — | 예 |
| B19 | if | `:286` `if current.ID != "" {` | `PositionID`, `firstNonEmpty` | — | 예 |
| B20 | if | `:318` `if result.OpenedInstance {` | — | — | 예 |
| B21 | if | `:319` `if _, err := tx.ExecContext(ctx, `` | `closedStamp`, `convergedState`, `fmt.Errorf`, `tx.ExecContext` | :327 | — |
| B22 | if | `:332` `if _, err := tx.ExecContext(ctx, `` | `fmt.Errorf`, `tx.ExecContext` | :341 | 예 |
| B23 | if | `:345` `if !result.OpenedInstance {` | — | — | 예 |
| B24 | if | `:346` `if _, err := tx.ExecContext(ctx, `` | `closeExitStateOnAdjustmentTx`, `closedStamp`, `convergedState`, `fmt.Errorf`, `tx.ExecContext` | :351 | — |
| B25 | if | `:357` `if err != nil {` | `scanPosition`, `tx.QueryRowContext` | :358 | 아니오 |
| B26 | if | `:363` `if err != nil {` | — | :364 | 아니오 |
| B27 | if | `:366` `if err := tx.Commit(); err != nil {` | `fmt.Errorf`, `tx.Commit` | :367, :371 | 예 |

## Calls and live bindings

`req.normalised` · `canonicalQuantity` · `adjustmentID` · `firstNonEmpty`(`:312`) · `convergedState` · `closeExitStateOnAdjustmentTx` · `tx.ExecContext`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`position_adjustments` 행 1개 · `positions`의 `quantity`/`avg_price`/`state`. **`exit_states`는 수량이 0이 될 때만 건드린다**(`closeExitStateOnAdjustmentTx` 첫 줄 `if !isZeroDecimal(newQuantity) { return false, nil }`).

## Safety conclusion

- **Safe edit boundary**: **여기가 a095의 신호원이다.** 수량이 늘어나는 조정은 오늘 `exit_states`를 전혀 건드리지 않는다 — 닫는 경로만 있고 다시 재는 경로가 없다. a095가 손절을 올린다면 그 신호는 이 함수가 만든 조정 행이며, **판단은 이 함수 안이 아니라 밖에서** 해야 한다(이 함수는 트랜잭션 안에서 원장 정합만 본다).
- **High-risk impact**: yes — 원장의 수량·평단 정본을 쓴다.
