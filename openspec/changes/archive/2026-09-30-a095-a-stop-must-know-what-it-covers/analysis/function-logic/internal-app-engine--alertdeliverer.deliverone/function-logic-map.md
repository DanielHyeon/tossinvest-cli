# Function Logic Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go` (`270`–`342`)
- Qualified: `alertDeliverer.deliverOne`
- AST evidence: `ast.json` (`source_sha256` df8a4171e8dced1b…)
- Risk scan: `risk-pattern-report.md`
- 분기 11 · return 4 · 호출 20

**역할.** PENDING 행 하나를 임차하고 보내며 결과를 원장에 정산한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `alert.Title` · `alert.Body` | 행에 저장된 문구 | 원장 `alert_outbox` | B9 창 — 그대로 `Publish` |
| `d.Publisher` | 전송기 | 배선 | B8 — nil이면 전송 수단 부재를 실패 시도로 셈 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:272` `if err != nil {` | `d.countRecordFailure`, `d.logf` | :276 | 예 |
| B2 | switch | `:278` `switch claim.Disposition {` | — | — | 예 |
| B3 | case | `:279` `case journal.ClaimAcquired:` | `d.forgetHeld` | — | 예 |
| B4 | case | `:283` `case journal.ClaimHeldElsewhere:` | — | — | 예 |
| B5 | if | `:292` `if d.reportHeld(alert.ID, claim.ExpiresAt) {` | `Format`, `claim.ExpiresAt.UTC`, `d.logf`, `d.reportHeld` | :297 | 예 |
| B6 | case | `:298` `default:` | `d.forgetHeld`, `d.forgetRecordRun` | :303 | 예 |
| B7 | if | `:305` `if claim.Stole {` | `d.logf` | — | 예 |
| B8 | if | `:316` `if d.Publisher == nil {` | `d.logf`, `errors.New` | — | 예 |
| B9 | else | `:326` `} else {` | `d.Publisher.Publish`, `obs.EventType` | — | 예 |
| B10 | if | `:333` `if perr != nil {` | `perr.Error` | — | 예 |
| B11 | if | `:337` `if perr != nil {` | `d.recordDelivery`, `d.recordFailedAttempt` | :339 | 예 |

## Calls and live bindings

`ClaimAlertByID` · `d.Publisher.Publish`(B9 창) · `d.recordDelivery` · `d.recordFailedAttempt`(B11).

결과값이 없다 — 오류를 돌려주지 않는다. claim 실패는 기록 실패 계수로(B1 창), 전송 실패는 원장의 실패 시도로(B11 창) 처리한다.

## State mutations and fallbacks

임차 · 정산 · 실패 계수.

## Safety conclusion

- **Safe edit boundary**: **10판 재추출 주석** — 분기 표 · 좌표 · 진입 실측은 현재 소스(아래 `source_sha256`)에서 기계로 다시 그렸다. 아래 산문은 3판(base `02716357`)의 판단이며, 그 뒤 a092 가 이 소스를 바꿨다(특히 `claimAndDeliver`는 이제 claim만 `n.mu` 아래에서 하고 전송은 잠금 밖 — `fbc6df5f`). a095 는 이 함수를 편집하지 않는다. 산문의 잠금 · 좌표 서술과 현재 소스가 어긋나면 표와 `review.md` §4.2 정정이 우선한다. **a095는 이 함수를 바꾸지 않는다.** B9 창이 보내는 것은 **행에 저장된** 제목 · 본문이다 — 발생 시점의 문구가 배달 시점까지 그대로 간다(r3 N3). B8은 전송기가 없을 때 실패 시도로 센다 — 알림 off 엔진에서 critical 행이 a124 정본의 지속 실패로 이어지는 경로이며, r3 N1(알림 켜짐을 critical의 전제로)의 근거다.
- **High-risk impact**: yes — critical 배달과 실패 판정의 자리다.
