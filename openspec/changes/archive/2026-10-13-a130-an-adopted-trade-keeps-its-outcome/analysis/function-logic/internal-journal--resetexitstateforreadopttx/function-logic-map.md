# Function Logic Map: `resetExitStateForReadoptTx`

- Source: `internal/journal/apply_hook.go` (편집 전 `699`–`745`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` 8daed316f087979f…) — branches 6 · returns 7
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 재편입 때 exit 상태를 새 t0 으로 재설정.
호출자: `ApplyPositionPolicy`(position_policy.go:145, ActionReadopt).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:702` `if err != nil {` | — | — | 무편집 갈래 |
| B2 | `:707` `if err := tx.QueryRowContext(ctx, `SELECT adoption_id,instance_seq FROM positions WHERE id` | — | — | 무편집 갈래 |
| B3 | `:715` `if err != nil {` | — | — | 무편집 갈래 |
| B4 | `:730` `if err != nil {` | — | — | 무편집 갈래 |
| B5 | `:734` `if err != nil {` | — | — | 무편집 갈래 |
| B6 | `:737` `if affected != 1 {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

**인용 근거만 · 무편집.** 근거 자리: UPDATE(:720-728)가 `entry_price` · `initial_risk` 를 다시 씀, 끝에 `ExitEventReadopted` exit 이벤트(:739-743) — a130 의 `readopted` 판정이 읽는 명시 기록.

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: no(무편집) — 인용 근거.
