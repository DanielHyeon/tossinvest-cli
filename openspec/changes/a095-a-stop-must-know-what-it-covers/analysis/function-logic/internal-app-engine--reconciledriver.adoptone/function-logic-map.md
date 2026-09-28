# Function Logic Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go` (`318`–`380`)
- Qualified: `ReconcileDriver.adoptOne`
- AST evidence: `ast.json` (`source_sha256` f121aba90cd05c31…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 3 · 호출 15

**역할.** 후보 하나를 편입한다 — 합성 손절 유도, 편입 기록 영속, exit state 개설, 편입 알림. 성공 여부를 bool로 답한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `exitpolicy.SyntheticStop`의 답 | 합성 손절 | 관측가 · `DefaultStopPct` | B1 — 실패면 `AdoptPosition`을 부르기 전에 false |
| `AdoptPosition`의 답 | 편입 기록 | 원장 | B2 — 거절 · 영속 실패면 false |
| `OpenAdoptedExitState`의 답 | exit state 개설 | 원장 | B3 — 실패해도 로그만 남기고 **true**를 반환 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:320` `if err != nil {` | `d.clk.Now`, `d.logDeferred`, `d.opts.Journal.AdoptPosition`, `err.Error`, `journal.RFC3339` | :322 | 아니오 |
| B2 | if | `:339` `if err != nil {` | `d.logDeferred`, `err.Error` | :341 | 아니오 |
| B3 | if | `:344` `if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != nil {` | `d.alert`, `d.label`, `d.logDeferred`, `d.opts.Journal.OpenAdoptedExitState`, `delete`, `err.Error`, `string` | :379 | 예 |

## Calls and live bindings

`exitpolicy.SyntheticStop`(B1 앞) · `d.opts.Journal.AdoptPosition`(B1 창) · `d.opts.Journal.OpenAdoptedExitState`(B3 조건) · `d.alert` · `delete(d.unmanaged, …)`(B3 창 — 반환 앞).

결과는 `bool`이다 — 오류를 돌려주지 않는다. 범주 ① ②는 `logDeferred` 로그 뒤 false, 범주 ③(B3 창)은 로그 뒤 true.

## State mutations and fallbacks

편입 기록 · exit state · 편입 알림 · 래치 해제(`d.unmanaged`).

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수의 본문을 바꾸지 않는다. 실패 세 범주를 경계로 기록한다(r4 R4-3, Manager 처분).** ① 편입 전 거절 — B1(합성 손절 유도 실패, `AdoptPosition` 호출 전) ② 영속 실패 — B2(`AdoptPosition` 거절 · 원장 검증 · 트랜잭션 실패) ③ 커밋 뒤 보호 미개설 — B3 창: exit state 개설이 실패해도 `true`. 「편입을 시도했으나 실패」(critical)는 ①② 즉 `false`이고, ③은 이 함수가 성공으로 답하므로 critical 요구 밖이다 — **이름 붙은 경계이며 그 자체가 후속 후보다**(`issues.md` I7). B1 · B2는 **미진입**이다.
- **High-risk impact**: yes — 편입의 쓰기 자리다.
