# Function Logic Map: `claimOwed`

- Source: `internal/journal/outbox.go` (`372`–`418`)
- Qualified: `claimOwed`
- AST evidence: `ast.json` (`source_sha256` ad74f6c9b295467e…)
- Risk scan: `risk-pattern-report.md`
- 분기 8 · return 7 · 호출 2

**역할.** 기존 outbox 행의 상태와 재알림 창으로 「이 관측이 전송을 빚졌나」와 「행을 새 발생으로 재무장하나」를 답한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | PENDING · DELIVERED · ACKNOWLEDGED · 기타 | 원장 | B2(PENDING) · B3(정착) · B8(기타) |
| `remindAfter` | 재알림 창 | `Notifier` 구성 | B4 — 0 이하이면 정착 행은 빚지지 않음 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | switch | `:378` `switch state {` | — | — | — |
| B2 | case | `:379` `case AlertPending:` | — | :381 | 예 |
| B3 | case | `:382` `case AlertDelivered, AlertAcknowledged:` | — | — | 예 |
| B4 | if | `:383` `if remindAfter <= 0 {` | `latestStamp` | :384 | 예 |
| B5 | if | `:387` `if !ok {` | `now.Sub` | :390 | 아니오 |
| B6 | if | `:393` `if elapsed < 0 {` | — | :405 | 예 |
| B7 | if | `:407` `if elapsed < remindAfter {` | — | :408, :410 | 예 |
| B8 | case | `:411` `default:` | — | :416 | 예 |

## Calls and live bindings

`latestStamp`(B3 창) · `now.Sub`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 순수 판정.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B2(`case AlertPending:`)의 창 return `:381`은 owed=참 · rearm=거짓이다 — PENDING 행은 재무장되지 않는다. 재무장은 정착 행이 창을 지났을 때(B7 창 `:410`)와 날짜를 매길 수 없는 · 미래 · 모르는 상태(B5 · B6 · B8)뿐이다. a089 R1의 「최신 발생 반영」은 a096~a099가 이 형태(창 뒤 재무장)로 대체했다(a089 archive `review.md:471`).
- **High-risk impact**: yes — 재알림 창의 판정이다.
