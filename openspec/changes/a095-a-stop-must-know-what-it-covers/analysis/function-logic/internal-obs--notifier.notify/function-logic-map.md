# Function Logic Map: `Notifier.Notify`

- Source: `internal/obs/notifier.go` (`107`–`116`)
- Qualified: `Notifier.Notify`
- AST evidence: `ast.json` (`source_sha256` d5b3004c638690fb…)
- Risk scan: `risk-pattern-report.md`
- 분기 1 · return 2 · 호출 4

**역할.** 이벤트를 등급 매기고 배달한다. **분기 하나가 durable 경로와 best-effort 경로를 가른다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `e` | 이벤트 | 호출자 | `SeverityOf(e.Type)`가 등급을 정한다 |
| `severity` | `SeverityOf`의 답 | 위 함수 | **critical이 아니면 B1이 best-effort로 보낸다** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:111` `if severity != SeverityCritical {` | `n.notifyCritical`, `n.publishBestEffort` | :113, :115 | 예 |

## Calls and live bindings

`SeverityOf` · `logEvent` · `publishBestEffort`(B1 안) · `notifyCritical`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 아래 두 경로가 각각 부작용을 갖는다. critical만 outbox 행을 만든다.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 주석이 반환 계약을 명시한다 — *"It returns an error only when a critical event could not be made durable."* 등급이 바뀌면 이 함수는 자동으로 다른 경로를 탄다.
- **High-risk impact**: yes — 알림이 원장에 남는지 아닌지가 여기서 갈린다.
