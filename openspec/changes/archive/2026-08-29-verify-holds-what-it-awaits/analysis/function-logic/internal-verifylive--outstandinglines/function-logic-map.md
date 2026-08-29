# Function Logic Map: `outstandingLines`

- Source: `internal/verifylive/record.go`
- Function: `internal/verifylive/record.go:outstandingLines`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

`Outstanding`의 본문에 **줄 index를 붙인 것**이다. 규칙은 그대로다 — 각 식별자의 마지막 비취소 언급이 이기고, 취소는 단조라서 나중 줄이 되살리지 못한다. index가 장식이 아닌 이유는 정리 규칙이 `HeldUntil`을 이 artifact에서 읽기 때문이다: 그것을 놓아줄 판정은 **그 필드를 실은 줄**과 비교되어야 하고 다른 줄과 비교되면 안 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | append-only 기록 | capability-verify*.jsonl | 빈 기록이면 nil |
| a.Cancelled | 취소가 확인된 객체 | 취소 성공 시에만 기록된다 | true면 결과에서 빠진다 |
| entry index i | 줄의 위치 | slice 순서 = 파일 순서 | append-only라 단조 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 497) — `for i, e := range entries {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `range` (line 498) — `for _, a := range e.Artifacts {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 500) — `if _, seen := latest[key]; !seen {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 504) — `if seen && prev.Cancelled && !a.Cancelled {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `range` (line 513) — `for _, key := range order {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `if` (line 514) — `if l := latest[key]; !l.Cancelled {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append` | ast.json calls (line 501) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — map과 slice를 지역에서 만들어 돌려준다.

## Safety conclusion

- Safe edit boundary: 신규이지만 로직은 `Outstanding`에서 그대로 옮겨왔다. 이동이 동작을 안 바꿨다는 것은 기존 `Outstanding` 테스트 전부가 통과하는 것으로 증명된다.
- High-risk impact: yes — 정리 대상과 노출 상한의 원천이다.
