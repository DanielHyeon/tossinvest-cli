# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (`288`–`361`)
- Qualified: `Notifier.claimAndDeliver`
- AST evidence: `ast.json` (`source_sha256` d705f78d68c1eff3…)
- Risk scan: `risk-pattern-report.md`
- 분기 7 · return 5 · 호출 18

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
| B1 | if | `:301` `if err != nil {` | `fmt.Sprintf` | — | 예 |
| B2 | if | `:314` `if n.Log != nil {` | `n.Log.Error`, `string` | — | 예 |
| B3 | if | `:317` `if n.Gate != nil {` | `n.Gate.Block`, `n.mu.Unlock` | :322 | 예 |
| B4 | switch | `:324` `switch claim.Disposition {` | — | — | 예 |
| B5 | case | `:325` `case journal.ClaimSettled:` | `n.mu.Unlock` | :334 | 예 |
| B6 | case | `:335` `case journal.ClaimHeldElsewhere:` | `n.deliver`, `n.logClaimHeld`, `n.logClaimStolen`, `n.mu.Unlock`, `string` | :347 | 예 |
| B7 | if | `:354` `if lost {` | — | :358, :360 | 예 |

## Calls and live bindings

`n.mu.Lock`/`n.mu.Unlock`(defer) · `n.Journal.ClaimAlertForDelivery` · `n.Gate.Block`(B3 창) · `n.logClaimHeld`(B6) · `n.deliver`(switch 뒤, 뮤텍스 안).

결과는 `(sent, owed, err)`다 — outbox claim 기록 실패만 오류로 돌려주고 그 전에 진입 게이트를 래치한다(B1 · B3 창). 전송 실패는 `deliver`가 처리하고 오류로 올리지 않는다.

## State mutations and fallbacks

outbox 행 claim · 실패 시 진입 게이트 래치(B3 창).

## Safety conclusion

- **Safe edit boundary**: **10판 재추출 주석** — 분기 표 · 좌표 · 진입 실측은 현재 소스(아래 `source_sha256`)에서 기계로 다시 그렸다. 아래 산문은 3판(base `02716357`)의 판단이며, 그 뒤 a092 가 이 소스를 바꿨다(특히 `claimAndDeliver`는 이제 claim만 `n.mu` 아래에서 하고 전송은 잠금 밖 — `fbc6df5f`). a095 는 이 함수를 편집하지 않는다. 산문의 잠금 · 좌표 서술과 현재 소스가 어긋나면 표와 `review.md` §4.2 정정이 우선한다. **a095는 이 함수를 바꾸지 않는다.** 두 사실이 3판의 제약이 된다. ① `n.deliver`가 `n.mu`를 잡은 채 불린다 — reconcile 자리의 critical 배달이 그동안 같은 `Notifier`의 다른 critical 호출자를 기다리게 한다(Q7). ② B5(`ClaimSettled`)가 같은 키의 재전송을 창 안에서 삼킨다 — 수량 증가 재알림이 critical이라면 키 설계가 이 창과 대조되어야 한다(Q4).
- **High-risk impact**: yes — 배달 직렬화와 재알림 억제의 자리다.
