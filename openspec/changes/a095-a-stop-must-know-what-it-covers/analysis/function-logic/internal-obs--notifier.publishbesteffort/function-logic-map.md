# Function Logic Map: `Notifier.publishBestEffort`

- Source: `internal/obs/notifier.go` (`138`–`150`)
- Qualified: `Notifier.publishBestEffort`
- AST evidence: `ast.json` (`source_sha256` d5b3004c638690fb…)
- Risk scan: `risk-pattern-report.md`
- 분기 2 · return 1 · 호출 6

**역할.** 일반 등급 알림을 보내고 **잊는다.** 이름이 계약이다 — outbox 행도 재시도도 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Publisher` | 전송기 | `Notifier` 구성 | **nil이면 B1이 조용히 반환한다 — 로그도 없다** |
| `n.Log` | 구조 로그 | `Notifier` 구성 | nil이면 전송 실패조차 기록되지 않는다(B2의 `&& n.Log != nil`) |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:139` `if n.Publisher == nil {` | — | :140 | 예 |
| B2 | if | `:142` `if err := n.Publisher.Publish(ctx, notificationFor(e, severity)); err != nil && n.Log != nil {` | `err.Error`, `n.Log.Warn`, `n.Publisher.Publish`, `notificationFor`, `string` | — | 예 |

## Calls and live bindings

`n.Publisher.Publish`(B2 조건 안) · `notificationFor` · `n.Log.Warn`(B2 안).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

외부 전송뿐. **원장에 아무것도 쓰지 않는다.**

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 이 경로가 조용한 것은 결함이 아니라 정의다 — 주석: *"treating its failure as an incident would make the grading meaningless."* 고칠 곳은 **무엇이 이 경로로 오느냐**이며 그것은 `criticalEvents`가 정한다.
- **High-risk impact**: no — 이 함수 자체는 안전하다. 위험한 것은 여기로 잘못 배달되는 이벤트다.
