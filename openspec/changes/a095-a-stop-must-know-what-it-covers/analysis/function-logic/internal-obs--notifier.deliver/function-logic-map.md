# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (`477`–`645`)
- Qualified: `Notifier.deliver`
- AST evidence: `ast.json` (`source_sha256` d705f78d68c1eff3…)
- Risk scan: `risk-pattern-report.md`
- 분기 27 · return 7 · 호출 42

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
| B1 | if | `:479` `if attempts <= 0 {` | `notificationFor` | — | 아니오 |
| B2 | for | `:485` `for attempt := 1; attempt <= attempts; attempt++ {` | — | — | 예 |
| B3 | if | `:486` `if n.Publisher == nil {` | `errors.New`, `n.Publisher.Publish` | — | 아니오 |
| B4 | if | `:491` `if err == nil {` | `n.Journal.MarkAlertDelivered` | — | 예 |
| B5 | if | `:493` `if markErr == nil {` | — | — | 예 |
| B6 | switch | `:494` `switch settled.Outcome {` | — | — | — |
| B7 | case | `:495` `case journal.SettleApplied:` | — | :496 | 예 |
| B8 | case | `:497` `case journal.SettleLeaseLost, journal.SettleAlreadySettled:` | `n.logLeaseLost` | :508 | 예 |
| B9 | case | `:509` `case journal.SettleNotFound:` | `errors.New` | — | 예 |
| B10 | case | `:511` `default:` | `fmt.Errorf`, `fmt.Sprintf`, `n.hook`, `n.readVerdict` | — | 아니오 |
| B11 | if | `:537` `if n.Log != nil {` | `err.Error`, `n.Journal.MarkAlertAttemptFailed`, `n.Log.Error`, `string` | :547 | 예 |
| B12 | if | `:551` `if markErr != nil {` | `n.Log.Error` | — | 예 |
| B13 | else | `:555` `} else if failed.Outcome != journal.SettleApplied {` | — | — | 예 |
| B14 | if | `:552` `if n.Log != nil {` | `n.Log.Error` | — | 예 |
| B15 | if | `:555` `} else if failed.Outcome != journal.SettleApplied {` | — | — | 예 |
| B16 | if | `:565` `if n.Log != nil {` | `n.Log.Error`, `n.logLeaseLost`, `string` | — | 예 |
| B17 | if | `:575` `if !isPreemption(failed.Outcome) && n.Gate != nil {` | `fmt.Sprintf`, `isPreemption`, `n.Gate.BlockUnlessClearedSince`, `n.hook`, `n.readVerdict` | :582 | 예 |
| B18 | if | `:584` `if attempt < attempts {` | — | — | 예 |
| B19 | if | `:585` `if !n.wait(ctx) {` | `n.Journal.ReleaseAlertClaim`, `n.hook`, `n.wait`, `relCancel`, `releaseCtx` | — | 예 |
| B20 | switch | `:603` `switch {` | — | — | — |
| B21 | case | `:604` `case relErr != nil:` | — | — | 아니오 |
| B22 | if | `:605` `if n.Log != nil {` | `n.Log.Error`, `string` | — | 아니오 |
| B23 | case | `:608` `case released.Outcome == journal.SettleApplied:` | — | — | 예 |
| B24 | case | `:611` `case !isPreemption(released.Outcome):` | `isPreemption`, `n.logLeaseLost` | — | 예 |
| B25 | if | `:616` `if n.Gate != nil {` | `fmt.Sprintf`, `n.Gate.BlockUnlessClearedSince`, `n.hook`, `n.readVerdict` | :622 | 예 |
| B26 | case | `:623` `default:` | `fmt.Sprintf`, `n.hook`, `n.logLeaseLost`, `n.readVerdict` | :633 | 아니오 |
| B27 | if | `:639` `if n.Log != nil {` | `n.Log.Error`, `string` | :644 | 예 |

## Calls and live bindings

`n.Publisher.Publish` · `n.Journal.MarkAlertDelivered` · `n.Journal.MarkAlertAttemptFailed` · `n.wait`(B20) · `n.Journal.ReleaseAlertClaim` · `n.Gate.Block`(B12 · B18 · B27 창).

결과는 `(sent, lost)`다 — 오류를 돌려주지 않는다. 실패는 로그 · 원장의 실패 시도 · 진입 게이트 래치(B12 · B18 · B27 창)로 처리한다.

## State mutations and fallbacks

outbox 행 상태 · 진입 게이트 래치.

## Safety conclusion

- **Safe edit boundary**: **10판 재추출 주석** — 분기 표 · 좌표 · 진입 실측은 현재 소스(아래 `source_sha256`)에서 기계로 다시 그렸다. 아래 산문은 3판(base `02716357`)의 판단이며, 그 뒤 a092 가 이 소스를 바꿨다(특히 `claimAndDeliver`는 이제 claim만 `n.mu` 아래에서 하고 전송은 잠금 밖 — `fbc6df5f`). a095 는 이 함수를 편집하지 않는다. 산문의 잠금 · 좌표 서술과 현재 소스가 어긋나면 표와 `review.md` §4.2 정정이 우선한다. **a095는 이 함수를 바꾸지 않는다.** B3(`n.Publisher == nil`)은 **미진입**이다 — 알림을 끈 엔진에서 critical이 어디로 가는지를 밟는 시험이 이 함수 단위로는 없다. 결정 (2)의 근거가 이 경로이므로 a095는 그 사실이 여기 오지 않음을 호출자 쪽 시험(2.5)으로 고정한다.
- **High-risk impact**: yes — 진입 차단 래치의 자리다.
