# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (`251`–`317`)
- Qualified: `Notifier.claimAndDeliver`
- AST evidence: `ast.json` (`source_sha256` 0bc75668ff17c3d6…)
- Risk scan: `risk-pattern-report.md`
- 분기 7 · return 5 · 호출 15

**역할.** outbox에 전송 의무를 묻고, 빚졌으면 배달한다 — 둘 다 `n.mu` 아래에서.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record.EventKey` | dedupe 키 | `n.eventKey(e)` — `e.Key`가 있으면 그것 | 같은 키의 전달된 행은 창 안에서 B5 |
| `n.remindAfter()` | 재알림 창 | `DefaultRemindAfter` | 창 안이면 `ClaimSettled` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:263` `if err != nil {` | `fmt.Sprintf` | — | 예 |
| B2 | if | `:276` `if n.Log != nil {` | `n.Log.Error`, `string` | — | 예 |
| B3 | if | `:279` `if n.Gate != nil {` | `n.Gate.Block` | :282 | 예 |
| B4 | switch | `:284` `switch claim.Disposition {` | — | — | 예 |
| B5 | case | `:285` `case journal.ClaimSettled:` | — | :293 | 예 |
| B6 | case | `:294` `case journal.ClaimHeldElsewhere:` | `n.deliver`, `n.logClaimHeld`, `n.logClaimStolen`, `string` | :305 | 예 |
| B7 | if | `:310` `if lost {` | — | :314, :316 | 예 |

## Calls and live bindings

`n.mu.Lock`/`n.mu.Unlock`(defer) · `n.Journal.ClaimAlertForDelivery` · `n.Gate.Block`(B3 창) · `n.logClaimHeld`(B6) · `n.deliver`(switch 뒤, 뮤텍스 안).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

outbox 행 claim · 실패 시 진입 게이트 래치(B3 창).

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 두 사실이 3판의 제약이 된다. ① `n.deliver`가 `n.mu`를 잡은 채 불린다 — reconcile 자리의 critical 배달이 그동안 같은 `Notifier`의 다른 critical 호출자를 기다리게 한다(Q7). ② B5(`ClaimSettled`)가 같은 키의 재전송을 창 안에서 삼킨다 — 수량 증가 재알림이 critical이라면 키 설계가 이 창과 대조되어야 한다(Q4).
- **High-risk impact**: yes — 배달 직렬화와 재알림 억제의 자리다.
