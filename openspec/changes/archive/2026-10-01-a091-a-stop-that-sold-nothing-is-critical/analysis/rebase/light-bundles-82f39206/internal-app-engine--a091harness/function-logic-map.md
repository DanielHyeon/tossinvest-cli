# Function Logic Map: `a091Harness`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go` (`108`–`127`)
- Qualified: `a091Harness`
- AST evidence: `ast.json` (`source_sha256` 5640f6fdbf0ded20…, **revision: base** — base `82f39206` 의 파일; 함수 본문은 HEAD 와 같고 창 안 커밋이 같은 파일 다른 자리를 고쳐 요구에 듦) — `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 2

**역할(경량 번들 — 시험 함수).** 하네스 (나) — 실제 RecordOnly + 원장 + 로그 캡처, 생산 모양(관측자 Log nil · ZeroFloorLog 전용 싱크).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 · 가짜 부품 | 시험 고정 | 시험 파일 | 단언 실패 → 시험 실패 |

## Branches and early returns

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:119` `if floor != nil {` |
| B2 | if | `:122` `if mutate != nil {` |

Exact AST return positions: `126:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(6). 생산 경로에 닿는 것은 하네스가 조립한 `ExitObserver` · 알림기 · 원장뿐 — 브로커 0.

## State mutations and fallbacks

임시 원장 · 버퍼만(시험 범위).

## Safety conclusion

- 생산 코드가 아니다 — 생산 동작 변화 0. 이 번들은 base 재고정 창의 증거 완결을 위한 경량 기록이다.
