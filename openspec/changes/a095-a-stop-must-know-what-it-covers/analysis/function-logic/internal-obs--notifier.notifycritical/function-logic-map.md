# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (`176`–`231`)
- Qualified: `Notifier.notifyCritical`
- AST evidence: `ast.json` (`source_sha256` 0bc75668ff17c3d6…)
- Risk scan: `risk-pattern-report.md`
- 분기 4 · return 3 · 호출 11

**역할.** critical 등급의 durable 경로. outbox 기록을 먼저 하고 배달하며, 배달이 빚졌는데 못 보냈으면 운영 모드를 승격한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | 원장 | `Notifier` 구성 | nil이면 B1 창에서 best-effort로 격하 |
| `sent, owed, err` | `claimAndDeliver`의 답 | 아래 번들 | B3 오류 → `n.escalate` · B4 `owed && !sent` → `n.escalate` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:177` `if n.Journal == nil {` | — | — | 예 |
| B2 | if | `:181` `if n.Log != nil {` | `encodeFields`, `n.Log.Warn`, `n.claimAndDeliver`, `n.eventKey`, `n.publishBestEffort`, `string` | :187 | 예 |
| B3 | if | `:201` `if err != nil {` | `fmt.Errorf`, `n.escalate` | :220 | 예 |
| B4 | if | `:223` `if owed && !sent {` | `n.escalate` | :230 | 예 |

## Calls and live bindings

`n.claimAndDeliver`(뮤텍스 안) · `n.escalate`(B3 창 · B4 창, 뮤텍스 밖) · `n.publishBestEffort`(B1 창).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

outbox 행(아래 번들) · 승격 기록(`escalate`).

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 결정 (2)는 운영자가 고른 상태에서 a095의 사실이 B3·B4의 `n.escalate`에 닿지 않을 것을 요구한다 — 그 사실이 이 함수에 **들어오지 않게** 하는 것이 해법이고, 들어온 뒤 거르는 분기를 이 함수에 더하지 않는다.
- **High-risk impact**: yes — 진입 차단 승격의 발화 자리다.
