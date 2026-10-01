# Function Logic Map: `TestA091TheAugustSecondReplay`

- Source: `internal/app/engine/a091_replay_test.go` (`104`–`135`)
- Qualified: `TestA091TheAugustSecondReplay`
- AST evidence: `ast.json` (`source_sha256` d9ca1043f02ee9e9…) — `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 4

**역할(경량 번들 — 시험 함수).** 8/2 재생 네 팔(생산 AlertDeliverer) — i2 에서 (i) · (iv) 의 미전달 줄 0 단언을 더함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 · 가짜 부품 | 시험 고정 | 시험 파일 | 단언 실패 → 시험 실패 |

## Branches and early returns

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:109` `if res.sends != 1 \|\| res.pendingNew != 0 \|\| res.latched \|\| res.mode != journal.ModeNormal \|\| res.undeliveredLines != 0 {` |
| B2 | if | `:117` `if res.pendingNew != 1 \|\| res.sends != 13 \|\| res.undeliveredLines != 1 \|\| !res.latched \|\| res.mode != journal.ModeEntryBlocked {` |
| B3 | if | `:124` `if res.pendingNew != 1 \|\| !res.latched \|\| res.mode != journal.ModeEntryBlocked \|\| res.undeliveredLines != 14 \|\| res.noPublisherLines != 13 {` |
| B4 | if | `:131` `if res.allRows != 0 \|\| res.sends != 0 \|\| res.latched \|\| res.mode != journal.ModeNormal \|\| res.undeliveredLines != 0 {` |

Exact AST return positions: none

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(16). 생산 경로에 닿는 것은 하네스가 조립한 `ExitObserver` · 알림기 · 원장뿐 — 브로커 0.

## State mutations and fallbacks

임시 원장 · 버퍼만(시험 범위).

## Safety conclusion

- 생산 코드가 아니다 — 생산 동작 변화 0. 이 번들은 base 재고정 창의 증거 완결을 위한 경량 기록이다.
