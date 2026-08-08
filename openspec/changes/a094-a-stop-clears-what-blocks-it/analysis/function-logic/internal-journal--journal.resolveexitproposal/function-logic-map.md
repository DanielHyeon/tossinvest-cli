# Function Logic Map: `Journal.ResolveExitProposal`

- Source: `internal/journal/apply_hook.go` (`810`–`869`)
- Qualified: `Journal.ResolveExitProposal`
- AST evidence: `ast.json` (`source_sha256` 88afd376da87b9b3…)
- Risk scan: `risk-pattern-report.md`
- 분기 14 · return 10 · 호출 23

**역할.** 무장된 발의를 해제한다. **a094 3판의 해동 기전이 부를 함수다.** `pending_action`을 NULL로 만드는 두 non-test writer 중 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | 대상 | 호출자 | B6이 행 부재를 가른다 |
| `resolution` | `ProposalRefused` 또는 `ProposalCancelled` | 호출자 | B2·B3·B4. 그 외는 `ErrInvalidRequest` |
| ``pending_action`` | 현재 무장 상태 | 원장 | **B8 `:842` — 이미 비었으면 무동작 반환**(멱등) |
| ``policy_kind`` | LADDER 여부 | 원장 | **B10 `:852` — LADDER면 rung을 되돌린다** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy,filldetect}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | switch | `:814` `switch resolution {` | — | — | — |
| B2 | case | `:815` `case ProposalRefused:` | — | — | 예 |
| B3 | case | `:816` `case ProposalCancelled:` | — | — | 예 |
| B4 | case | `:818` `default:` | `fmt.Errorf`, `j.db.BeginTx`, `j.nowString`, `string` | :819 | 아니오 |
| B5 | if | `:825` `if err != nil {` | `Scan`, `fmt.Errorf`, `tx.QueryRowContext`, `tx.Rollback` | :826 | 아니오 |
| B6 | if | `:836` `if errors.Is(err, sql.ErrNoRows) {` | `errors.Is`, `fmt.Errorf` | :837 | 아니오 |
| B7 | if | `:839` `if err != nil {` | `fmt.Errorf` | :840 | 예 |
| B8 | if | `:842` `if strings.TrimSpace(pendingAction.String) == "" {` | `strings.TrimSpace` | :843 | 예 |
| B9 | if | `:846` `if _, err := tx.ExecContext(ctx, `` | `fmt.Errorf`, `tx.ExecContext` | :850 | 예 |
| B10 | if | `:852` `if kind == ExitPolicyLadder {` | — | — | 예 |
| B11 | if | `:853` `if rung, err := exitpolicy.RungIndex(strings.TrimSpace(pendingLevel.String)); err == nil {` | `exitpolicy.RungIndex`, `strings.TrimSpace` | — | 예 |
| B12 | if | `:854` `if err := rollBackRungTx(ctx, tx, id, rung-1, now); err != nil {` | `rollBackRungTx` | :855 | 아니오 |
| B13 | if | `:859` `if err := appendExitEventTx(ctx, tx, exitEventRow{` | `appendExitEventTx`, `strings.TrimSpace` | :863 | 예 |
| B14 | if | `:865` `if err := tx.Commit(); err != nil {` | `fmt.Errorf`, `tx.Commit` | :866, :868 | 예 |

## Calls and live bindings

`tx.QueryRowContext` · `tx.ExecContext`(`:846` NULL 쓰기) · `exitpolicy.RungIndex`(B11) · `rollBackRungTx`(B12) · `appendExitEventTx`(B13) · `tx.Commit`(B14).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`pending_action`·`pending_level`·`pending_intent_id` → NULL(`:846-849`) · LADDER면 `active_rung`를 한 칸 되돌림(`exit_state.go:980-987`) · `exit_events` 1행.

## Safety conclusion

- **Safe edit boundary**: **a094 3판이 새 호출자를 더한다** — attempt가 비-CONFIRMED 종결에 이르면 그 intent를 가리키는 발의를 해제한다. **함수 본문은 바꾸지 않는다.** 안전 확인 둘: (1) `rollBackRungTx`는 `active_rung`만 쓰고 **손절 가격은 건드리지 않는다**(§6 무관), (2) B8이 멱등을 보장하므로 중복 호출이 무해하다. `RungIndex`는 음수 label을 거부하므로(`ladder.go:536-538`) `pending_level="-1"`인 행은 rung 되돌림 없이 해제된다.
- **High-risk impact**: yes — 발의 해제는 다음 주기의 재제출을 허용한다.
