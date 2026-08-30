# Function Logic Map: `decidedAfter`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:decidedAfter`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-reopens-conditional-chain`

이 change가 추가한 leaf. 단계 `id`의 최신 기록 항목이 artifact `a`를 처음 기록한 항목보다 뒤에 있는지 답한다. 시계가 아니라 append-only 기록의 색인 순서를 쓴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | 기록 순서 | 증거 기록 | 빈 경우 두 색인 모두 -1 → false |
| id StepID | catalogue 단계 | `StepConditionalCancel`만 호출됨 | 항목 없으면 false |
| a Artifact | `Outstanding`이 돌려준 것 | 같은 기록 | 생성 항목 미발견이면 false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 152) — `for i := range entries {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `range` (line 153) — `for _, x := range entries[i].Artifacts {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 154) — `if x.Kind == a.Kind && x.ID == a.ID {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 159) — `if created >= 0 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `if` (line 163) — `if created < 0 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `range` (line 167) — `for i := range entries {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `if` (line 168) — `if entries[i].StepID == id {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 없음 — 순수 함수. 지역 `created`·`decided`만 쓴다.

## Safety conclusion

- Safe edit boundary: 신규 함수라 기존 동작이 없다. fail-closed 기본값(false = 정리하지 않음)이 경계다.
- High-risk impact: yes — 라이브 취소 대상 선정에 쓰인다. 방향은 취소를 덜 보내는 쪽이다.
