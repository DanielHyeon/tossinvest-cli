# Function Logic Map: `resetExitStateForReadoptTx`

- Source: `internal/journal/apply_hook.go` (`684`–`730`)
- Qualified: `resetExitStateForReadoptTx`
- AST evidence: `ast.json` (`source_sha256` 88afd376da87b9b3…)
- Risk scan: `risk-pattern-report.md`
- 분기 6 · return 7 · 호출 17

**역할.** 재편입 시 보호 기준을 **다시 세우는 유일한 쓰기 자리**다. 주석이 그렇게 선언한다 — *"the only reset writer for the four guarded columns"*.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | 대상 | 호출자 | B6이 정확히 1행을 요구한다 |
| `generation/instance_seq` | 세대 | 원장 | B2가 대조 |
| `새 entry/stop/risk` | 새 기준 | `positionpolicy.ActionReadopt` | **운영자 행동으로만 온다** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:687` `if err != nil {` | `fmt.Errorf` | :688 | 아니오 |
| B2 | if | `:692` `if err := tx.QueryRowContext(ctx, `SELECT adoption_id,instance_seq FROM positions WHERE id=?`,` | `Scan`, `fmt.Errorf`, `policyKindForID`, `seedPolicyIdentity`, `strings.TrimSpace`, `tx.QueryRowContext` | :694 | — |
| B3 | if | `:700` `if err != nil {` | `nullableString`, `string`, `tx.ExecContext` | :701 | 아니오 |
| B4 | if | `:715` `if err != nil {` | `fmt.Errorf`, `result.RowsAffected` | :716 | 아니오 |
| B5 | if | `:719` `if err != nil {` | — | :720 | 아니오 |
| B6 | if | `:722` `if affected != 1 {` | `appendExitEventTx`, `fmt.Errorf`, `string` | :723, :725 | 예 |

## Calls and live bindings

`tx.QueryRowContext`(B2) · `tx.ExecContext`(UPDATE) · `RowsAffected`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`exit_states`의 `policy_kind`·`policy_id`·`entry_price`·`initial_stop`·`initial_risk`를 덮어쓴다.

## Safety conclusion

- **Safe edit boundary**: **a095가 기대는 사실**: 기준을 다시 세우는 경로가 이것 **하나뿐**이고, 그것이 운영자 행동(`positionpolicy.ActionReadopt`)에서만 불린다. 즉 앱에서 추가 매수해도 이 함수는 **절대 불리지 않는다.** a095가 손절을 올리는 경로를 만든다면 그것은 이 함수의 형제여야 하고 같은 「유일 쓰기 자리」 규율을 따라야 한다.
- **High-risk impact**: yes — 손절선을 덮어쓰는 유일한 자리다.
