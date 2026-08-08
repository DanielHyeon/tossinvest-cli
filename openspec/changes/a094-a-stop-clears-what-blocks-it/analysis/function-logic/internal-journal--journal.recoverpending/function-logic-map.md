# Function Logic Map: `Journal.RecoverPending`

- Source: `internal/journal/recovery.go` (`86`–`125`)
- Qualified: `Journal.RecoverPending`
- AST evidence: `ast.json` (`source_sha256` bc81bfd5a3654ad9…)
- Risk scan: `risk-pattern-report.md`
- 분기 10 · return 6 · 호출 13

**역할.** 재시작 시 미종결 attempt를 재시작 규칙으로 정리한다. **세션 중에 부르면 원장을 위조한다** — 3판도 이 판단을 유지한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `PendingAttempts` | 미종결 attempt 전수 | 원장 | B2가 순회 |
| `rec.State` | attempt 상태 | 원장 | **B4·B6·B9가 갈린다** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy,filldetect}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:88` `if err != nil {` | — | :89 | 아니오 |
| B2 | range | `:93` `for _, rec := range pending {` | — | — | 예 |
| B3 | switch | `:94` `switch rec.State {` | — | — | — |
| B4 | case | `:95` `case StateRecorded:` | `j.handleFor` | — | 예 |
| B5 | if | `:97` `if err := handle.Settle(ctx, StateNotDispatched, ReasonRestartNotDispatched,` | `append`, `fmt.Errorf`, `handle.Settle` | :99 | — |
| B6 | case | `:103` `case StateDispatchStarted:` | `j.handleFor` | — | 예 |
| B7 | if | `:105` `if err := handle.MarkInDoubt(ctx, ReasonRestartInDoubt,` | `append`, `fmt.Errorf`, `handle.MarkInDoubt`, `j.blockedFor` | :107 | — |
| B8 | if | `:111` `if err != nil {` | `append` | :112 | 아니오 |
| B9 | case | `:116` `case StateAcked, StateInDoubt:` | `j.blockedFor` | — | 예 |
| B10 | if | `:118` `if err != nil {` | `append` | :119, :124 | 아니오 |

## Calls and live bindings

`j.PendingAttempts` · `j.handleFor` · `handle.Settle`(B5) · `handle.MarkInDoubt`(B7) · `j.blockedFor`(B8·B10).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

**B4 `:95` `StateRecorded` → `Settle(NOT_DISPATCHED, "found at startup with no dispatch recorded")` · B6 `:103` `StateDispatchStarted` → `MarkInDoubt("process stopped after dispatch started")`.** 세션 중이면 둘 다 거짓 사유다. **B9 `:116` `Acked`·`InDoubt`는 읽기만 한다.**

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않고 세션 중에 부르지도 않는다(SHALL NOT).** 3판의 해동 경로는 B9가 다루는 상태(`IN_DOUBT`)만 대상으로 하며, 그 상태에 대해 이 함수가 하는 일은 `blockedFor` 읽기뿐이다 — **즉 3판이 필요로 하는 것은 이 함수가 아니라 그 뒤의 해소다.**
- **High-risk impact**: yes — attempt 상태를 종결시킨다.
