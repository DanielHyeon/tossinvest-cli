# Function Logic Map: `armExitProposalTx`

- Source: `internal/journal/apply_hook.go` (`655`–`677`)
- Qualified: `armExitProposalTx`
- AST evidence: `ast.json` (`source_sha256` 88afd376da87b9b3…)
- Risk scan: `risk-pattern-report.md`
- 분기 4 · return 5 · 호출 10

**역할.** 발의를 무장한다. **B3 `:666`이 두 번째 발의를 거부한다** — a094 3판이 R2의 자기 방향 검사를 철회하는 근거다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | 대상 | 호출자 | B1이 행 부재를 가른다 |
| ``exit_states.pending_action`` | 이미 무장됐는가 | 원장 | **B3 — 비어 있지 않으면 `ErrProposalPending`으로 거부** |
| `proposal` | 무장할 발의 | `ExecutableProposal()` | B4의 UPDATE 값 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy,filldetect}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:660` `if errors.Is(err, sql.ErrNoRows) {` | `errors.Is`, `fmt.Errorf` | :661 | 아니오 |
| B2 | if | `:663` `if err != nil {` | `fmt.Errorf` | :664 | 예 |
| B3 | if | `:666` `if strings.TrimSpace(action.String) != "" {` | `fmt.Errorf`, `strings.TrimSpace` | :667 | 예 |
| B4 | if | `:669` `if _, err := tx.ExecContext(ctx, `` | `fmt.Errorf`, `nullableString`, `tx.ExecContext` | :674, :676 | 예 |

## Calls and live bindings

`tx.QueryRowContext` · `tx.ExecContext`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`exit_states`의 `pending_action`·`pending_level`·`pending_intent_id`를 쓴다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** 주석이 논거를 갖는다 — *"A second proposal while one is outstanding is refused rather than overwritten."* **엔진 자신의 초과 매도는 여기서 이미 막힌다**(3판 R2 축소의 근거). 이 함수가 막지 못하는 것은 사용자가 앱에 넣은 매도이며, 그것을 막으려던 검사가 손절을 영구 보류시켰다.
- **High-risk impact**: yes — 초과 매도의 1차 방벽이다.
