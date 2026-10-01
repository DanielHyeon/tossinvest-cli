# Function Logic Map: `TestA091AZeroHoldingIsNotAFailedStop`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go` (`341`–`356`)
- Qualified: `TestA091AZeroHoldingIsNotAFailedStop`
- AST evidence: `ast.json` (`source_sha256` 9e8cdf8722000cda…) — `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 3

**역할(경량 번들 — 시험 함수).** 보유 0(Holdings 한정 0)은 옛 종류 normal — i2 에서 payload cause `no_holding` 단언을 더함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 · 가짜 부품 | 시험 고정 | 시험 파일 | 단언 실패 → 시험 실패 |

## Branches and early returns

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:345` `if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 0 {` |
| B2 | if | `:349` `if len(lines) != 1 \|\| !strings.Contains(fmt.Sprint(lines[0]["detail"]), "계좌에 보유가 없다") {` |
| B3 | if | `:353` `if a != "no_holding" {` |

Exact AST return positions: none

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(14). 생산 경로에 닿는 것은 하네스가 조립한 `ExitObserver` · 알림기 · 원장뿐 — 브로커 0.

## State mutations and fallbacks

임시 원장 · 버퍼만(시험 범위).

## Safety conclusion

- 생산 코드가 아니다 — 생산 동작 변화 0. 이 번들은 base 재고정 창의 증거 완결을 위한 경량 기록이다.
