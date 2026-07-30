# Function Logic Map: `dependsOn`

- Source: `internal/verifylive/redo.go`
- Function: `internal/verifylive/redo.go:dependsOn`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-reopens-conditional-chain`

이 change가 추가한 leaf. `Step.DependsOn`에 주어진 id가 있는지 답한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| s Step | catalogue 단계 | `Steps()` | `DependsOn` nil이면 false |
| id StepID | 의존 대상 | 호출자 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 139) — `for _, d := range s.DependsOn {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `if` (line 140) — `if d == id {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 없음 — 순수 함수.

## Safety conclusion

- Safe edit boundary: 신규 함수. catalogue 데이터만 읽는다.
- High-risk impact: no — 데이터 조회. 위험은 호출자(`subjectLost`)에 있다.
