# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (`420`–`574`)
- Qualified: `Notifier.deliver`
- AST evidence: `ast.json` (`source_sha256` 0bc75668ff17c3d6…)
- Risk scan: `risk-pattern-report.md`
- 분기 27 · return 6 · 호출 30

**역할.** 빚진 critical 알림을 재시도 예산 안에서 보내고, 끝내 못 보내면 진입 게이트를 래치한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Publisher` | 전송기 | `Notifier` 구성 | nil이면 B3 창에서 `lastErr`를 세우고 루프를 벗어난다 |
| `attempts` | 재시도 횟수 | `n.Attempts` 또는 기본값(B1) | B19 · B20 사이 대기 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:422` `if attempts <= 0 {` | `notificationFor` | — | 아니오 |
| B2 | for | `:428` `for attempt := 1; attempt <= attempts; attempt++ {` | — | — | 예 |
| B3 | if | `:429` `if n.Publisher == nil {` | `errors.New`, `n.Publisher.Publish` | — | 아니오 |
| B4 | if | `:434` `if err == nil {` | `n.Journal.MarkAlertDelivered` | — | 예 |
| B5 | if | `:436` `if markErr == nil {` | — | — | 예 |
| B6 | switch | `:437` `switch settled.Outcome {` | — | — | — |
| B7 | case | `:438` `case journal.SettleApplied:` | — | :439 | 예 |
| B8 | case | `:440` `case journal.SettleLeaseLost, journal.SettleAlreadySettled:` | `n.logLeaseLost` | :451 | 아니오 |
| B9 | case | `:452` `case journal.SettleNotFound:` | `errors.New` | — | 아니오 |
| B10 | case | `:454` `default:` | `fmt.Errorf`, `fmt.Sprintf` | — | 아니오 |
| B11 | if | `:478` `if n.Log != nil {` | `n.Log.Error`, `string` | — | 예 |
| B12 | if | `:483` `if n.Gate != nil {` | `err.Error`, `n.Gate.Block`, `n.Journal.MarkAlertAttemptFailed` | :491 | 예 |
| B13 | if | `:495` `if markErr != nil {` | `n.Log.Error` | — | 예 |
| B14 | else | `:499` `} else if failed.Outcome != journal.SettleApplied {` | — | — | 예 |
| B15 | if | `:496` `if n.Log != nil {` | `n.Log.Error` | — | 예 |
| B16 | if | `:499` `} else if failed.Outcome != journal.SettleApplied {` | — | — | 예 |
| B17 | if | `:509` `if n.Log != nil {` | `n.Log.Error`, `n.logLeaseLost`, `string` | — | 예 |
| B18 | if | `:519` `if failed.Outcome == journal.SettleNotFound && n.Gate != nil {` | `fmt.Sprintf`, `n.Gate.Block` | :523 | 예 |
| B19 | if | `:525` `if attempt < attempts {` | — | — | 예 |
| B20 | if | `:526` `if !n.wait(ctx) {` | `n.Journal.ReleaseAlertClaim`, `n.wait`, `relCancel`, `releaseCtx` | — | 예 |
| B21 | switch | `:543` `switch {` | — | — | — |
| B22 | case | `:544` `case relErr != nil:` | — | — | 아니오 |
| B23 | if | `:545` `if n.Log != nil {` | `n.Log.Error`, `string` | — | 아니오 |
| B24 | case | `:548` `case released.Outcome == journal.SettleApplied:` | — | — | 예 |
| B25 | case | `:551` `default:` | `fmt.Sprintf`, `n.logLeaseLost` | :561 | 아니오 |
| B26 | if | `:565` `if n.Log != nil {` | `n.Log.Error`, `string` | — | 예 |
| B27 | if | `:570` `if n.Gate != nil {` | `n.Gate.Block` | :573 | 예 |

## Calls and live bindings

`n.Publisher.Publish` · `n.Journal.MarkAlertDelivered` · `n.Journal.MarkAlertAttemptFailed` · `n.wait`(B20) · `n.Journal.ReleaseAlertClaim` · `n.Gate.Block`(B12 · B18 · B27 창).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

outbox 행 상태 · 진입 게이트 래치.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B3(`n.Publisher == nil`)은 **미진입**이다 — 알림을 끈 엔진에서 critical이 어디로 가는지를 밟는 시험이 이 함수 단위로는 없다. 결정 (2)의 근거가 이 경로이므로 a095는 그 사실이 여기 오지 않음을 호출자 쪽 시험(2.5)으로 고정한다.
- **High-risk impact**: yes — 진입 차단 래치의 자리다.
