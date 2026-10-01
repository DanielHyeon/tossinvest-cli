# Function Logic Map: `TestA091TheProductionAssemblyReadsTheLoadedSwitch`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go` (`872`–`901`)
- Qualified: `TestA091TheProductionAssemblyReadsTheLoadedSwitch`
- AST evidence: `ast.json` (`source_sha256` 9e8cdf8722000cda…) — `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 6

**역할(경량 번들 — 시험 함수).** 생산 배선이 로드된 notifications.enabled 와 엔진 로거로 덮음 — i2 에서 조립 로거를 덮지 않고 `eng.Log != nil` 단언.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 · 가짜 부품 | 시험 고정 | 시험 파일 | 단언 실패 → 시험 실패 |

## Branches and early returns

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | range | `:873` `for _, loaded := range []bool{true, false} {` |
| B2 | if | `:882` `if err != nil {` |
| B3 | if | `:885` `if eng.Log == nil {` |
| B4 | if | `:890` `if err != nil {` |
| B5 | if | `:893` `if got := observer.OptionsForTest().ZeroFloorLog; got == nil \|\| got != eng.Log {` |
| B6 | if | `:896` `if got := observer.OptionsForTest().NotificationsEnabled; got != loaded {` |

Exact AST return positions: none

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(18). 생산 경로에 닿는 것은 하네스가 조립한 `ExitObserver` · 알림기 · 원장뿐 — 브로커 0.

## State mutations and fallbacks

임시 원장 · 버퍼만(시험 범위).

## Safety conclusion

- 생산 코드가 아니다 — 생산 동작 변화 0. 이 번들은 base 재고정 창의 증거 완결을 위한 경량 기록이다.
