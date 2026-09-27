# Function Logic Map: `ExitObserver.alertUnmanaged`

- Source: `internal/app/engine/exitloop.go` (`1601`–`1618`)
- Qualified: `ExitObserver.alertUnmanaged`
- AST evidence: `ast.json` (`source_sha256` 522d5d81c4992c57…)
- Risk scan: `risk-pattern-report.md`
- 분기 1 · return 1 · 호출 3

**역할.** exit 관측이 본 무관리 보유를 알린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.unmanaged[p.ID]` | 이미 알렸나 | 프로세스 메모리 map | B1 창의 return |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:1602` `if o.unmanaged[p.ID] {` | `o.alert`, `o.label`, `string` | :1603 | 예 |

## Calls and live bindings

`o.alert` · `o.label` · `string`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`o.unmanaged[p.ID] = true`(메모리) · 알림 1건. 키 `exit.position_unmanaged|<posID>` — reconcile 쪽 `alertUnmanaged`와 **같은 철자**다.

## Safety conclusion

- **Safe edit boundary**: **본문은 바꾸지 않는다.** 결정 (1)에 따라 이 자리의 사실은 normal로 남고, 결정 (3)(iii)에 따라 키가 reconcile 자리와 달라야 한다.
- **High-risk impact**: yes — exit goroutine 안의 발신이다.
