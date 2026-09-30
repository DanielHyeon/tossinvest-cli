# Function Logic Map: `resetExitStateForReadoptTx`

- Source: `internal/journal/apply_hook.go` (`699`–`745`)
- Qualified: `resetExitStateForReadoptTx`
- AST evidence: `ast.json` (`source_sha256` 459f1791fa91a6c7…)
- Risk scan: `risk-pattern-report.md`
- 분기 6 · return 7 · 호출 17

**역할.** 재편입 시 보호 기준 전체를 새 관측으로 다시 세운다. 주석: *"the only reset writer for the four guarded execution-time columns"* — 네 열은 손절 열이 아니다(보이스 A A-4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `observation` | 재편입 관측가 · 합성 손절 | `positionpolicy.ActionReadopt` | B1 — 무효면 거절 |
| `positionID` | 대상 | 호출자 | B6 — 정확히 1행 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:702` `if err != nil {` | `fmt.Errorf` | :703 | 아니오 |
| B2 | if | `:707` `if err := tx.QueryRowContext(ctx, `SELECT adoption_id,instance_seq FROM positions WHERE id=?`,` | `Scan`, `fmt.Errorf`, `policyKindForID`, `seedPolicyIdentity`, `strings.TrimSpace`, `tx.QueryRowContext` | :709 | — |
| B3 | if | `:715` `if err != nil {` | `nullableString`, `string`, `tx.ExecContext` | :716 | 아니오 |
| B4 | if | `:730` `if err != nil {` | `fmt.Errorf`, `result.RowsAffected` | :731 | 아니오 |
| B5 | if | `:734` `if err != nil {` | — | :735 | 아니오 |
| B6 | if | `:737` `if affected != 1 {` | `appendExitEventTx`, `fmt.Errorf`, `string` | :738, :740 | 예 |

## Calls and live bindings

`exitpolicy.OpenRatchetState` · `tx.QueryRowContext`(B2) · `tx.ExecContext`(UPDATE) · `appendExitEventTx`.

결과에 `error`가 있다 — 원장(트랜잭션 · 질의) 호출의 오류와 입력 검증 실패를 되던진다(위 표의 return 열이 그 자리다). 브로커 호출은 없다(9판 r8 N5 — 값 단위 재전수).

## State mutations and fallbacks

`exit_states`의 `entry_price` · `initial_stop` · `initial_risk` · `baseline_price` · `high_water` 등을 새 값으로 덮어쓴다.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않고 부르지도 않는다.** 분기 여섯은 전부 오류·행 수 검사이며 **이전 `baseline_price`와 비교하는 분기가 없다** — 이 경로는 기준선을 낮출 수 있다(운영자 행동에서만 불린다). 2판의 「기준을 다시 세우는 유일한 쓰기 자리」는 「reset 경로가 하나」로만 참이다.
- **High-risk impact**: yes — 손절선을 덮어쓴다.
