# Function Logic Map: `stepRun.markHeld`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:stepRun.markHeld`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

이 객체가 **일부러 살아 있고 누가 기다리는지**를 한 번에 기록한다. `markDeliberate`의 대체다. 필드가 둘로 나뉘는 것은 읽는 쪽이 둘이기 때문이다 — `Deliberate`는 사람에게('이건 누수가 아니다', 종료 검사와 모든 화면이 읽는다), `HeldUntil`은 정리 규칙에게('누가 판정할 때까지 손대지 마라'). **이 단계가 실제로 기록한 artifact에만 닿는다.** 객체를 만들지 않고 읽기만 하는 단계(오늘은 `conditional-persist`가 그렇다)는 표시할 것이 없어 호출이 무해한 no-op이 되며, 그것이 붙잡음을 선언하는 두 번째 경로가 생기지 않게 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| kind, id string | 표시할 객체 | 호출 단계 | 일치하는 artifact가 없으면 no-op |
| gate StepID | 놓아줄 단계 | 호출부 상수 | 카탈로그 밖 값은 영구 붙잡음 — AST 가드가 막는다 |
| chain string | 측정 사슬 | Runner.chainOf 또는 새 토큰 | 빈 값이면 사슬 없음 |
| note string | 사람이 읽을 사유 | 호출부 리터럴 | 덮어쓴다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 740) — `for i := range sr.artifacts {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 741) — `if sr.artifacts[i].Kind == kind && sr.artifacts[i].ID == id && !sr.artifacts[i].Cancelled {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- `sr.artifacts[i]`의 `Deliberate`·`HeldUntil`·`ChainID`·`Note`를 제자리에서 바꾼다. 취소된 artifact는 건드리지 않는다(조건이 `!Cancelled`).

## Safety conclusion

- Safe edit boundary: 필드 둘을 더 쓰는 것과 gate·chain 인자를 받는 것. 순회 조건과 `Deliberate`·`Note` 동작은 무변경.
- High-risk impact: yes — 여기서 찍힌 gate가 그 객체의 취소 시점을 정한다.
