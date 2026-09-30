# Function Logic Map: `EntryGate.ProjectOperatingMode`

- Source: `internal/execgw/modegate.go`
- AST evidence: `ast.json` — **편집 전**, :35–51, 분기 3 · 반환 1 · 호출 5, source_sha256 `029961389516…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): AC2 — 한 번의 `g.mu` 구간 안에서 지우고 다시 넣는다(오늘은 `:37` delete → `:38` unlock → `:50` `g.Block` 재잠금 — 빈 창). 울타리: 마지막으로 적용한 순번보다 큰 투영만 적용(초기값 0, 「보다 큰」). 상태 세대(revision)는 모드 사유 **존재**가 바뀔 때만 +1(C20). 해제 세대(clearEpochs)는 건드리지 않는다(a124).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !rec.BlocksEntry()` (:40) | — | — | (미실행) |
| B2 | `if rec.Actor != ""` (:44) | — | — | `TestARejectedCredentialTightensTheOperatingMode`, `TestTheModeLatchIsReplacedRatherThanAccumulated` |
| B3 | `if rec.Cause != ""` (:47) | — | — | `TestARejectedCredentialTightensTheOperatingMode`, `TestTheModeLatchIsReplacedRatherThanAccumulated` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `delete`, `g.Block`, `g.mu.Lock`, `g.mu.Unlock`, `rec.BlocksEntry` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: High-risk(진입 게이트).
