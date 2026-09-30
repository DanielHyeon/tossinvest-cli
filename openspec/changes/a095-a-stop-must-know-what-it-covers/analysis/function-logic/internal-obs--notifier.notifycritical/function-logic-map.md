# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (`195`–`250`)
- Qualified: `Notifier.notifyCritical`
- AST evidence: `ast.json` (`source_sha256` d705f78d68c1eff3…)
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
| B1 | if | `:196` `if n.Journal == nil {` | — | — | 예 |
| B2 | if | `:200` `if n.Log != nil {` | `encodeFields`, `n.Log.Warn`, `n.claimAndDeliver`, `n.eventKey`, `n.publishBestEffort`, `string` | :206 | 예 |
| B3 | if | `:220` `if err != nil {` | `fmt.Errorf`, `n.escalate` | :239 | 예 |
| B4 | if | `:242` `if owed && !sent {` | `n.judge` | :249 | 예 |

## Calls and live bindings

`n.claimAndDeliver`(뮤텍스 안) · `n.escalate`(B3 창 · B4 창, 뮤텍스 밖) · `n.publishBestEffort`(B1 창).

결과는 `error`다 — `claimAndDeliver`의 기록 실패만 돌려주고, 그 전에 `escalate`를 시도한다(B3 창). 전송 실패는 오류가 아니다(B4 창에서 승격).

## State mutations and fallbacks

outbox 행(아래 번들) · 승격 기록(`escalate`).

## Safety conclusion

- **Safe edit boundary**: **10판 재추출 주석** — 분기 표 · 좌표 · 진입 실측은 현재 소스(아래 `source_sha256`)에서 기계로 다시 그렸다. 아래 산문은 3판(base `02716357`)의 판단이며, 그 뒤 a092 가 이 소스를 바꿨다(특히 `claimAndDeliver`는 이제 claim만 `n.mu` 아래에서 하고 전송은 잠금 밖 — `fbc6df5f`). a095 는 이 함수를 편집하지 않는다. 산문의 잠금 · 좌표 서술과 현재 소스가 어긋나면 표와 `review.md` §4.2 정정이 우선한다. **a095는 이 함수를 바꾸지 않는다.** 결정 (2)는 운영자가 고른 상태에서 a095의 사실이 B3·B4의 `n.escalate`에 닿지 않을 것을 요구한다 — 그 사실이 이 함수에 **들어오지 않게** 하는 것이 해법이고, 들어온 뒤 거르는 분기를 이 함수에 더하지 않는다.
- **High-risk impact**: yes — 진입 차단 승격의 발화 자리다.
