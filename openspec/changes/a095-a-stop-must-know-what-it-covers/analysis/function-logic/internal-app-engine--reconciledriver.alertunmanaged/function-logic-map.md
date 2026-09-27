# Function Logic Map: `ReconcileDriver.alertUnmanaged`

- Source: `internal/app/engine/adoption.go` (`392`–`432`)
- Qualified: `ReconcileDriver.alertUnmanaged`
- AST evidence: `ast.json` (`source_sha256` f121aba90cd05c31…)
- Risk scan: `risk-pattern-report.md`
- 분기 6 · return 1 · 호출 5

**역할.** 엔진이 관리하지 않는 보유를 알린다. why-matrix(B2 switch)가 사유를 고른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.unmanaged[p.ID]` | 이미 알렸나 | 프로세스 메모리 map | B1 — 프로세스당 1회 |
| `d.opts.Adoption` | 편입 설정 | 런타임 config | B3~B6이 사유 문구를 고른다 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:393` `if d.unmanaged[p.ID] {` | — | :394 | 예 |
| B2 | switch | `:404` `switch {` | — | — | — |
| B3 | case | `:405` `case d.opts.Adoption.Rejected != "":` | — | — | 예 |
| B4 | case | `:407` `case d.opts.Adoption.Excludes(p.Symbol):` | `d.opts.Adoption.Excludes` | — | 예 |
| B5 | case | `:409` `case d.opts.Adoption.Enabled:` | — | — | 예 |
| B6 | case | `:412` `case d.opts.Adoption.Included(p.Symbol):` | `d.alert`, `d.label`, `d.opts.Adoption.Included`, `string` | — | 예 |

## Calls and live bindings

`d.opts.Adoption.Excludes`(B4) · `d.opts.Adoption.Included`(B6) · `d.alert` · `d.label`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`d.unmanaged[p.ID] = true`(메모리) · 알림 1건. 키는 `exit.position_unmanaged|<posID>` — exit 관측 자리와 **같은 철자**다.

## Safety conclusion

- **Safe edit boundary**: **사유 분기는 이미 사실별로 갈려 있다** — B3(설정 거부) · B4(exclude) · B5(enabled 시도 실패) · B6(include 지정 시도 실패) · 기본(off∧미지정). 결정 (2)는 B4와 기본을 normal로 두라 하고, B5는 운영자가 고른 상태가 아니다. B3 · B6의 분류는 결정이 덮지 않는다(Q2). 이 함수의 키는 결정 (3)(iii)에 따라 exit 관측 자리와 갈라야 한다.
- **High-risk impact**: yes — reconcile 쪽 무보호 보고의 자리다.
