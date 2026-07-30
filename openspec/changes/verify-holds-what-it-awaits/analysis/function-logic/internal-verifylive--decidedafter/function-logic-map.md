# Function Logic Map: `decidedAfter`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:decidedAfter`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

> **base revision** — 이 함수는 이 change가 수정하지 않았다. diff 문맥에만 걸렸고, AST는 `base-commit.txt`의 소스에서 뽑았다.

**이 change가 삭제한 함수다.** base revision으로 고정한다. `heldAfter`가 대체했다. 하던 일은 같은 계열이다 — gate 단계의 마지막 줄이 artifact보다 뒤에 있는가. 다른 점은 기준선이 artifact의 **최초** 언급이었다는 것이고, 그래서 지목이 나중에 바뀌는 경우 (같은 객체를 다시 붙잡는 경우)를 표현할 수 없었다. design.md D2가 두 정의를 대조한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | append-only 기록 | capability-verify*.jsonl | 빈 기록이면 false |
| id StepID | gate 단계 — 호출부가 `conditional-cancel`로 고정 | cleanupFrom | 줄이 없으면 -1 → false |
| a Artifact | 판정 대상 객체 | Outstanding | 최초 언급을 못 찾으면 false(fail-closed) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 152) — `for i := range entries {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `range` (line 153) — `for _, x := range entries[i].Artifacts {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 154) — `if x.Kind == a.Kind && x.ID == a.ID {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 159) — `if created >= 0 {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `if` (line 163) — `if created < 0 {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `range` (line 167) — `for i := range entries {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B7 | `if` (line 168) — `if entries[i].StepID == id {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 두 index를 세어 비교한다.

## Safety conclusion

- Safe edit boundary: 무변경(base revision). 삭제 자체가 이 change의 편집이며, 대체 함수가 같은 fail-closed 방향을 유지한다.
- High-risk impact: yes — 삭제 전까지 조건주문 취소 여부를 정하던 술어다.
