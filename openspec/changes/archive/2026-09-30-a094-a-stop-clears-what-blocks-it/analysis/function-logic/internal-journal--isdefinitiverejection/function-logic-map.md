# Function Logic Map: `isDefinitiveRejection`

- Source: `internal/journal/dispatch.go` (`349`–`356`)
- Qualified: `isDefinitiveRejection`
- AST evidence: `ast.json` (`source_sha256` 42165f21f0707b4d…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 2 · 호출 0

**역할.** HTTP status 하나를 받아 '요청 자체를 서술하는 거절인가'를 답한다. 순수 함수다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `statusCode` | 브로커가 준 HTTP status | `ClassifyHTTPMutation` B5의 인자 | 없음 — 순수 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | switch | `:350` `switch statusCode {` | — | — | — |
| B2 | case | `:351` `case 400, 401, 403, 404, 405, 415, 422:` | — | :352 | 예 |
| B3 | case | `:353` `default:` | — | :354 | 예 |

## Calls and live bindings

없다 (`ast.json`의 `calls`가 `null`).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다. 순수 함수이고 부작용이 없다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** 409를 B2 목록에 넣는 것은 `request-in-progress`까지 확정 거절로 만들어 살아 있는 주문을 은퇴시킨다(design D1·D6).
- **High-risk impact**: yes — 이 함수의 답이 attempt의 종결/동결을 가른다.
