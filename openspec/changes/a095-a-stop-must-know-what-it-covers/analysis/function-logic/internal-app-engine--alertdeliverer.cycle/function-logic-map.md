# Function Logic Map: `alertDeliverer.cycle`

- Source: `internal/app/engine/alertdelivery.go` (`235`–`264`)
- Qualified: `alertDeliverer.cycle`
- AST evidence: `ast.json` (`source_sha256` 5791a31af9d24079…)
- Risk scan: `risk-pattern-report.md`
- 분기 4 · return 3 · 호출 12

**역할.** 배달 실행자의 한 사이클 — PENDING 행을 한도 아래 먼저 골라 행마다 `deliverOne`을 부른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `PendingAlertsForDelivery`의 답 | PENDING 행 | 원장 | B1 — 나열 실패는 지속 실패로 셈 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:243` `if err != nil {` | `d.countListFailure`, `fmt.Errorf` | :246 | 예 |
| B2 | if | `:249` `if len(pending) < d.batch() {` | `d.batch`, `d.pruneRecordRuns`, `len` | — | 예 |
| B3 | range | `:254` `for _, alert := range pending {` | — | — | 예 |
| B4 | if | `:258` `if ctx.Err() != nil {` | `ctx.Err`, `d.deliverOne` | :259, :263 | 예 |

## Calls and live bindings

`d.led().PendingAlertsForDelivery` · `d.countListFailure`(B1) · `d.pruneRecordRuns`(B2) · `d.deliverOne`(B4 창).

결과는 `error` 하나다 — 나열 실패(B1)를 지속 실패로 센 뒤 오류로 돌려준다. 행별 배달 실패는 `deliverOne`이 원장에 기록하고 여기로 돌려주지 않는다.

## State mutations and fallbacks

실패 계수 · 행 배달(`deliverOne` 경유).

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 이 실행자는 PENDING 행을 원장에서 읽어 보낸다 — 사실이 해소됐는지 묻지 않는다(r3 N3).
- **High-risk impact**: yes — critical 배달의 실행자다.
