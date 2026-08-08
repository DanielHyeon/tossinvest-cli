# Function Logic Map: `SeverityOf`

- Source: `internal/obs/event.go` (`309`–`314`)
- Qualified: `SeverityOf`
- AST evidence: `ast.json` (`source_sha256` acf37eb4e4529612…)
- Risk scan: `risk-pattern-report.md`
- 분기 1 · return 2 · 호출 0

**역할.** 이벤트 종류 하나를 받아 등급을 답한다. **`criticalEvents` map만 본다** — 순수 함수이고 분기가 하나다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | `EventType` 문자열 | 호출자 | 없음 — 순수 |
| `criticalEvents` | 등급 표(`event.go:279-298`, 18종) | **소스의 리터럴 map** | 미등재는 조용히 `SeverityNormal` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:310` `if criticalEvents[t] {` | — | :311, :313 | 예 |

## Calls and live bindings

없다.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다.

## Safety conclusion

- **Safe edit boundary**: **a095가 바꾸는 것은 이 함수가 아니라 그것이 읽는 map이다.** 함수 본문은 그대로 두고 `EventExitPositionUnmanaged`를 `criticalEvents`에 등재한다. 주석이 그 설계를 명시한다 — *"Genuinely critical conditions are named in the table above."*
- **High-risk impact**: yes — 이 답이 알림의 durable 여부를 정한다.
