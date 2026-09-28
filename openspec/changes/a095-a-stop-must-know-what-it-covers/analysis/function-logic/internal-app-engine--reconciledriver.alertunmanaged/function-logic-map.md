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
| `d.unmanaged[p.ID]` | 이미 알렸나 | 프로세스 메모리 map — 키는 포지션 id뿐 | B1 — 프로세스당 1회. 사실 · 등급을 보지 않는다 |
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

결과값이 없다 — 오류를 돌려주지 않는다. `d.alert`(→ `ReconcileDriver.alert` B2)가 `Notify`의 오류를 로그로만 남긴다.

## State mutations and fallbacks

`d.unmanaged[p.ID] = true`(메모리) · 알림 1건. 키는 `exit.position_unmanaged|<posID>` — exit 관측 자리와 **같은 철자**다.

## Safety conclusion

- **Safe edit boundary**: **사유 분기는 이미 사실별로 갈려 있다** — B3(설정 거부) · B4(exclude) · B5(enabled 시도 실패) · B6(include 지정 시도 실패) · 기본(off∧미지정). 결정 (2)는 B4와 기본을 normal로 두라 하고, B5는 운영자가 고른 상태가 아니다. B3 · B6의 분류는 결정이 덮지 않는다(Q2). 이 함수의 키는 결정 (3)(iii)에 따라 exit 관측 자리와 갈라야 한다. **5판: B1 래치가 편집 경계 안이다(r4 R4-1, Manager 처분).** 오늘 B1은 포지션 id만으로, `d.alert`(→ `Notify`) **앞**에서 억제해 앞 사이클의 normal 보고(예: 연기)가 뒤 사이클의 critical 사실(시도 실패)을 삼키고, 기록이 실패해도 다음 주기 재시도를 막는다. **6판 원칙(r5 R5-1)**: 메모리 래치는 **normal 보고 전용**이고 critical 보고는 래치를 무조건 지난다 — 중복은 outbox 키와 정본 재알림 창이 맡는다. normal 래치의 키는 사실 식별자(R5-2)다 — design D1 「후보별 결과와 억제 키」.
- **High-risk impact**: yes — reconcile 쪽 무보호 보고의 자리다.
