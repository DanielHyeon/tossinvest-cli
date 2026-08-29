# Function Logic Map: `stepRun.markDeliberate`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:stepRun.markDeliberate`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

> **base revision** — 이 함수는 이 change가 수정하지 않았다. diff 문맥에만 걸렸고, AST는 `base-commit.txt`의 소스에서 뽑았다.

**이 change가 이름과 계약을 바꾼 함수다.** base revision으로 고정한다. `markHeld`가 대체했다. 하던 일은 '이 객체는 실수가 아니다'를 화면에 말하는 것 하나였다. 정리 규칙은 이 표시를 읽지 않았고 kind로 갈라진 자기 규칙을 따로 갖고 있었다 — 그것이 이 change가 닫는 간극이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| kind, id string | 표시할 객체 | 호출 단계 | 일치하는 artifact가 없으면 아무 일도 없다 |
| note string | 사람이 읽을 사유 | 호출부 리터럴 | 덮어쓴다 |
| sr.artifacts | 이 단계가 기록한 객체들 | sr.created/cancelled | 비어 있으면 no-op |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 731) — `for i := range sr.artifacts {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 732) — `if sr.artifacts[i].Kind == kind && sr.artifacts[i].ID == id && !sr.artifacts[i].Cancelled {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- `sr.artifacts[i]`의 `Deliberate`와 `Note`를 제자리에서 바꾼다.

## Safety conclusion

- Safe edit boundary: 무변경(base revision). 대체 함수가 같은 두 필드를 같은 조건에서 쓰고 두 필드를 더 쓴다.
- High-risk impact: no — 표시만 했다. 이 함수가 정리 판정에 관여하지 않았다는 것이 이 change의 출발점이다.
