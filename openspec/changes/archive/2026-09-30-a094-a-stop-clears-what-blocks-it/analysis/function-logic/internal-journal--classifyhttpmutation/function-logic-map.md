# Function Logic Map: `ClassifyHTTPMutation`

- Source: `internal/journal/dispatch.go` (`301`–`345`)
- Qualified: `ClassifyHTTPMutation`
- AST evidence: `ast.json` (`source_sha256` 42165f21f0707b4d…)
- Risk scan: `risk-pattern-report.md`
- 분기 7 · return 6 · 호출 10

**역할.** 전송 상태와 HTTP status를 dispatch 분류로 바꾼다. 의도적으로 비관적이다 — 증명되지 않은 것은 전부 Ambiguous.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `send` | `SendNotStarted` 등 전송 진행도 | dispatch 추적기 | B2에서 NotSent를 가른다 |
| `statusCode` | 브로커 status. 0은 '상태 없음' | `statusOf` | B6이 0을 따로 잡는다 |
| `err` | 전송 오류 | 호출자 | B1이 먼저 본다 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:302` `if err != nil {` | — | — | 예 |
| B2 | if | `:303` `if send == SendNotStarted {` | `err.Error`, `fmt.Sprintf`, `send.String` | :304, :311 | 예 |
| B3 | switch | `:320` `switch {` | — | — | 예 |
| B4 | case | `:321` `case statusCode >= 200 && statusCode < 300:` | — | :322 | 예 |
| B5 | case | `:323` `case isDefinitiveRejection(statusCode):` | `fmt.Errorf`, `fmt.Sprintf`, `isDefinitiveRejection` | :324 | 예 |
| B6 | case | `:330` `case statusCode == 0:` | `errors.New` | :331 | 예 |
| B7 | case | `:337` `default:` | `fmt.Errorf`, `fmt.Sprintf` | :338 | 예 |

## Calls and live bindings

`isDefinitiveRejection`(B5) · `fmt.Sprintf` · `errors.New`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다. `DispatchOutcome` 값을 만들어 돌려줄 뿐이다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** 409가 B7 `default:`로 가는 것은 status만 보면 옳은 판단이다. a094는 이 함수가 **불리기 전에** code로 분류한다(`classifyMutation` B3).
- **High-risk impact**: yes — 여기서 나온 분류가 IN_DOUBT 여부를 정한다.
