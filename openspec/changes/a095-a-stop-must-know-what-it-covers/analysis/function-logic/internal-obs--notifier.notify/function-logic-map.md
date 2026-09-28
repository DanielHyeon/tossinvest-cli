# Function Logic Map: `Notifier.Notify`

- Source: `internal/obs/notifier.go` (`130`–`139`)
- Qualified: `Notifier.Notify`
- AST evidence: `ast.json` (`source_sha256` 0bc75668ff17c3d6…)
- Risk scan: `risk-pattern-report.md`
- 분기 1 · return 2 · 호출 4

**역할.** 이벤트를 등급 매기고 배달한다. B1이 best-effort 경로와 durable 경로를 가른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `e` | 이벤트 | 호출자 | `SeverityOf(e.Type)`가 등급을 정한다 |
| `severity` | `SeverityOf`의 답 | 위 함수 | critical이 아니면 B1 창의 `n.publishBestEffort` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:134` `if severity != SeverityCritical {` | `n.notifyCritical`, `n.publishBestEffort` | :136, :138 | 예 |

## Calls and live bindings

`SeverityOf` · `n.logEvent`(등급 판정 **앞**, 두 경로 공통) · `n.publishBestEffort`(B1 창) · `n.notifyCritical`(B1 뒤).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 두 경로가 각자 부작용을 갖는다. `n.mu`는 이 함수가 잡지 않는다(`claimAndDeliver` 번들 참조).

## Safety conclusion

- **Safe edit boundary**: **편집 경계는 Q1의 답에 달렸다(4판, r3 N4)** — (a)면 본문 불변, (b)면 `SeverityOf` 계약 변경과 함께 경계 재선언. 어느 답이든 결정 (1)의 요구 「exit goroutine 에 critical Notify 를 새로 두지 않는다」는 exit 관측 자리의 사실이 B1 창(`publishBestEffort`)으로 가는 것으로 성립한다 — 그 경로에는 `n.mu`도 outbox도 재시도 대기도 없다(네트워크 발행 1회는 동기다).
- **High-risk impact**: yes — 알림이 원장에 남는지, 진입 차단에 닿는지가 여기서 갈린다.
