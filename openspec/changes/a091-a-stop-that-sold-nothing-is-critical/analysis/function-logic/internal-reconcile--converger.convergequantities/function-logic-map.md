# Function Logic Map: `Converger.ConvergeQuantities`

- Source: `internal/reconcile/converge.go` (`142`–`273`)
- Qualified: `Converger.ConvergeQuantities`
- AST evidence: `ast.json` (`source_sha256` 00a784b3f6b3a3ef…) — 편집 뒤 `3ec1efd2` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 15 · 반환 10

**편집.** 편집하지 않는다 — M1(Manager 지시): 포지션 수량의 생산자. 수량 불일치에서 원장 투영을 계좌 값으로 맞춘다.

**역할.** 불일치마다 `ApplyPositionAdjustment(NewQuantity: mismatch.Authority())`(`:218` — `Authority()` = `m.Broker`, `compare.go:266`). 0 으로 수렴하면 exit 상태를 닫는다(B13).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `mismatch.Broker` | 계좌가 보고한 수량(철자는 대사 비교가 만든 값) | `reconcile.Compare` | — |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `3ec1efd2` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:144` `if len(diff.Quantities) == 0 {` | 예 |
| B2 | if | `:147` `if c == nil \|\| c.Journal == nil {` | 예 |
| B3 | if | `:152` `if account == "" {` | 아니오 |
| B4 | if | `:156` `if asOf == "" {` | 아니오 |
| B5 | if | `:173` `if len(credited) == 0 {` | 예 |
| B6 | if | `:178` `if c.Credit != nil {` | 예 |
| B7 | if | `:194` `if err != nil {` | 아니오 |
| B8 | range | `:197` `for _, mismatch := range diff.Quantities {` | 예 |
| B9 | if | `:200` `if reason != "" {` | 예 |
| B10 | if | `:206` `if err != nil {` | 아니오 |
| B11 | if | `:228` `if errors.Is(err, journal.ErrAdjustmentStale) {` | 예 |
| B12 | if | `:233` `if err != nil {` | 예 |
| B13 | if | `:244` `if result.ClosedExitState {` | 예 |
| B14 | if | `:251` `if c.Alert == nil {` | 아니오 |
| B15 | if | `:254` `if err := c.Alert.ManagedPositionClosedExternally(ctx, ManagedCloseAlert{` | 예 |

Exact AST return positions: `145:3`, `148:3`, `153:3`, `157:3`, `174:4`, `195:3`, `207:4`, `229:4`, `234:4`, `272:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(31).

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `c.Journal.ApplyPositionAdjustment` | `:172` | 투영 수렴 | 원장 트랜잭션 |

## State mutations and fallbacks

포지션 수량 · exit 상태(0 수렴 시 종료).

## Safety conclusion

- 이 값의 철자는 0 판정에 닿지 않는다: exit 루프는 `isZeroQuantity(p.Quantity)`(수치) 포지션을 건너뛰고(`exitloop.go:541`), 나머지는 `canonicalSnapshotContext` 가 양수 강제 + `RatString` 으로 바꾼다.
