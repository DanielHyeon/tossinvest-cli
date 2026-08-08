# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go` (`117`–`204`)
- Qualified: `Journal.OpenExitState`
- AST evidence: `ast.json` (`source_sha256` 99c96bdba4a08f7c…)
- Risk scan: `risk-pattern-report.md`
- 분기 20 · return 15 · 호출 33

**역할.** 포지션의 보호 상태를 만든다. **`entry_price`·`initial_stop`·`initial_risk`가 여기서 얼어붙는다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `seed.EntryPrice` | 진입가 | 호출자 | **주석은 *"the position's cost basis"*라 쓴다** — 그러나 평단이 바뀌어도 갱신되지 않는다 |
| `seed.InitialStop` | 진입 결정의 손절가 | 호출자 | 둘의 차가 R 분모 |
| `seed.PolicyKind/PolicyID` | RATCHET 또는 LADDER | config | 정책 정체성 검증 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:119` `if id == "" {` | `fmt.Errorf`, `strings.TrimSpace` | :120 | 아니오 |
| B2 | if | `:123` `if kind == "" {` | — | — | 예 |
| B3 | if | `:126` `if kind != ExitPolicyRatchet && kind != ExitPolicyLadder {` | `fmt.Errorf`, `strings.TrimSpace` | :127 | 예 |
| B4 | switch | `:131` `switch kind {` | — | — | — |
| B5 | case | `:132` `case ExitPolicyRatchet:` | — | — | 예 |
| B6 | if | `:133` `if policyID != "" {` | `fmt.Errorf` | :134 | 아니오 |
| B7 | case | `:136` `case ExitPolicyLadder:` | — | — | 예 |
| B8 | if | `:137` `if policyID == "" {` | — | — | 예 |
| B9 | else | `:139` `} else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" {` | — | — | 예 |
| B10 | if | `:139` `} else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" {` | `exitpolicy.CommonPolicyByID`, `exitpolicy.OpenRatchetState`, `fmt.Errorf` | :140 | 예 |
| B11 | if | `:147` `if err != nil {` | `fmt.Errorf`, `j.db.BeginTx`, `j.nowString` | :148 | 예 |
| B12 | if | `:153` `if err != nil {` | `Scan`, `fmt.Errorf`, `tx.QueryRowContext`, `tx.Rollback` | :154 | 아니오 |
| B13 | if | `:163` `if errors.Is(err, sql.ErrNoRows) {` | `errors.Is`, `fmt.Errorf` | :164 | 아니오 |
| B14 | if | `:166` `if err != nil {` | `fmt.Errorf` | :167 | 예 |
| B15 | if | `:169` `if !position.ExitEligible(decisionID.String, adoptionID.String) {` | `fmt.Errorf`, `position.ExitEligible`, `seedPolicyIdentity`, `strings.TrimSpace` | :170 | 예 |
| B16 | if | `:174` `if err != nil {` | — | :175 | 예 |
| B17 | if | `:178` `if _, err := tx.ExecContext(ctx, `` | `string`, `tx.ExecContext` | — | 예 |
| B18 | if | `:187` `if isUniqueViolation(err) {` | `fmt.Errorf`, `isUniqueViolation` | :188, :190 | 예 |
| B19 | if | `:192` `if err := appendExitEventTx(ctx, tx, exitEventRow{` | `appendExitEventTx`, `string` | :196 | 예 |
| B20 | if | `:198` `if err := tx.Commit(); err != nil {` | `fmt.Errorf`, `j.ExitState`, `tx.Commit` | :199, :203 | 예 |

## Calls and live bindings

`strings.TrimSpace` · `exitpolicy` 정책 검증 · `j.db` INSERT.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`exit_states` 행 1개를 만든다. `baseline_price`에 진입 손절을 함께 넣어 **보호 없는 순간이 없게** 한다.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수의 t0 동결을 바꾸지 않는다.** 여기서 쓴 값을 나중에 낮추는 경로를 만드는 것이 위험한 방향이다. a095는 **높이는 방향만** 다루고, 그것도 이 함수가 아니라 별도 경로에서 한다.
- **High-risk impact**: yes — R의 분모와 손절선의 원천.
