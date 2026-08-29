# Function Logic Map: `heldAfter`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:heldAfter`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

붙잡음이 **풀렸는지** 판정한다. `decidedAfter`의 대체이고 차이는 기준선 하나다: `Outstanding`이 고른 그 줄, 즉 **gate를 지목한 바로 그 줄**과 비교한다. 읽는 줄과 재는 줄을 같게 두는 것이 정의의 전부이며, 그래야 취소가 실패한 뒤 다른 단계가 객체를 다시 언급하기만 해도 붙잡히는 일(M22의 재발)이 정의 수준에서 불가능해진다. **시계가 아니라 index다** — 기록은 append-only라 index만 단조이고, 취소 줄은 zero time을 싣는다. 시각 기반 해제는 발동 대기 창 한복판에서 발화하는데, 대기가 길다는 것은 실패가 아니라 시장이 아직 안 움직였다는 뜻이다(design.md D3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | append-only 기록 | capability-verify*.jsonl | gate 줄이 없으면 decided=-1 |
| gate StepID | 지목된 단계 | holdGate | 카탈로그에 없는 값이면 영원히 안 풀린다 — AST 가드가 막는다 |
| at int | gate를 지목한 줄의 index | outstandingLines | 그 줄보다 앞선 판정은 놓아주지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 185) — `for i := range entries {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 186) — `if entries[i].StepID == gate {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — index 두 개를 비교한다.

## Safety conclusion

- Safe edit boundary: 신규 leaf. fail-closed 방향(`decided`가 -1이면 false)은 `decidedAfter`와 동일하다.
- High-risk impact: yes — 이 술어가 true면 살아 있는 브로커 객체가 취소 승인 목록에 오른다.
