# Function Logic Map: `subjectLost`

- Source: `internal/verifylive/redo.go`
- Function: `internal/verifylive/redo.go:subjectLost`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-reopens-conditional-chain`

이 change가 추가한 leaf. 통과했지만 그 통과가 남긴 조건주문이 사라졌고, 그것을 필요로 하는 단계가 아직 통과하지 못한 경우를 답한다. 세 조건이 모두 AND다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | 기록 | 증거 기록 | 빈 경우 `Outstanding` 빈 목록, `Passed` false → 의존 단계 있으면 true |
| step Step | catalogue 단계 | `Steps()` | `DependsOn` 역참조에 쓰인다 |
| newest Entry | 그 단계의 최신 항목 | `LastEntry` | `pass`가 아니면 즉시 false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 109) — `if newest.Verdict != VerdictPass {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `range` (line 113) — `for _, a := range newest.Artifacts {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 114) — `if a.Kind == KindConditional && a.Deliberate {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 119) — `if !left {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `range` (line 122) — `for _, a := range Outstanding(entries) {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `if` (line 123) — `if a.Kind == KindConditional {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `range` (line 127) — `for _, dep := range Steps() {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B8 | `if` (line 128) — `if dep.Deferred != "" \|\| !dependsOn(dep, step.ID) {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B9 | `if` (line 131) — `if !Passed(entries, dep.ID) {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Outstanding` | ast.json calls (line 122) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `Steps` | ast.json calls (line 127) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `dependsOn` | ast.json calls (line 128) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `Passed` | ast.json calls (line 131) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 없음 — 순수 함수. 지역 `left`만 쓴다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 세 조건 중 하나라도 거짓이면 false이므로 기본 방향은 '아무것도 보내지 않는다'.
- High-risk impact: yes — true면 조건주문 등록이 재제안될 수 있다. 승인 게이트는 무변경.
