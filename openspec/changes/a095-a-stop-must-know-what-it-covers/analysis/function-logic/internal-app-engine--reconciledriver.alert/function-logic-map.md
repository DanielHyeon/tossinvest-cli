# Function Logic Map: `ReconcileDriver.alert`

- Source: `internal/app/engine/reconcileloop.go` (`552`–`560`)
- Qualified: `ReconcileDriver.alert`
- AST evidence: `ast.json` (`source_sha256` 50a2c0f0b133fc0a…)
- Risk scan: `risk-pattern-report.md`
- 분기 2 · return 1 · 호출 3

**역할.** 대사 루프의 알림 발신 — `Notify`를 부르고, 오류는 로그로만 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.opts.Alerts` | 알림기 | 배선 | B1 — nil이면 return |
| `Notify`의 오류 | critical의 durable 기록 실패 | `Notifier` | B2 — 로그만 남기고 호출자에 돌려주지 않음 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:553` `if d.opts.Alerts == nil {` | — | :554 | 아니오 |
| B2 | if | `:556` `if err := d.opts.Alerts.Notify(ctx, e); err != nil && d.opts.Log != nil {` | `d.opts.Alerts.Notify`, `d.opts.Log.Error`, `string` | — | 예 |

## Calls and live bindings

`d.opts.Alerts.Notify`(B2 조건) · `d.opts.Log.Error`(B2 창).

이 함수는 결과값이 없고 오류를 **되던지지 않는다** — B2가 `Notify`의 오류를 로그로만 남긴다. 호출자는 기록 실패를 알 수 없다(7판 r6 R6-2).

## State mutations and fallbacks

없다 — 알림 발신과 로그.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B2가 `Notify`의 오류(critical outbox 기록 실패)를 로그로만 남기므로 호출자 `alertUnmanaged`는 기록이 실패했는지 모른다. 오늘은 B1 래치가 `d.alert` 앞에 걸려 있어 기록 실패가 다음 주기에 다시 시도되지 않는다(r5 R5-1). 6판 원칙 — critical은 래치를 지나므로 다음 관측이 다시 기록을 시도한다 — 은 이 함수를 바꾸지 않고 성립한다.
- **High-risk impact**: yes — 대사 쪽 critical 발신의 통로다.
