# Function Logic Map: `Batch.Verify`

- Source: `internal/verifylive/confirm.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| input | 사람이 타이핑한 문자열 | TTY | 공백 제거 후 정확 일치만 통과 |
| now | 호출자의 시계 | runner.now | 만료가 우선 판정된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if b.Expired(now) {` (internal/verifylive/confirm.go:254, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestConfirmBatchRejectsAnythingButTheNonce, TestConfirmBatchExpires, TestExpiredIsTheSameWindowVerifyUses |
| B2 | `if strings.TrimSpace(input) != b.Nonce {` (internal/verifylive/confirm.go:257, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestConfirmBatchRejectsAnythingButTheNonce, TestConfirmBatchExpires, TestExpiredIsTheSameWindowVerifyUses |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Batch.Expired | 창 판정을 위임 — 두 번째 시계 규칙을 만들지 않는다 | 없음 | AST callees |
| strings.TrimSpace | 입력 정규화 | 없음 | AST callees |

## State mutations and fallbacks

- 없음 — 순수 판정.
- 동작 변경 없음: 만료 판정을 Expired로 추출했을 뿐 결과는 동일하다.

## Safety conclusion

- Safe edit boundary: 만료 우선·정확 일치. 대소문자 무시나 접두 일치는 승인의 의미를 바꾼다.
- High-risk impact: no — TTY 경로의 판정. 콘솔은 더 이상 이 함수를 쓰지 않는다.
