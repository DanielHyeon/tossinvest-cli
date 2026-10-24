# Function Logic Map: `classifyCancel`

- Source: `internal/flatten/flatten.go` (454-474)
- Qualified function: `classifyCancel`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 1f66eed56b478d18…)
- AST evidence: `ast.json` — AST branches 6, 반환 5, 호출 2
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

게이트웨이 결과 → flatten step. B3(458) FAILED_CONFIRMED → FlattenStepFailed, B5(464) IN_DOUBT → InDoubt.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| out | Outcome | Gateway.Cancel | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 457:3, 463:3, 465:3, 471:3, 473:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `switch at 455:2` | switch out.State { | 아래 「State mutations」 | 소스 줄 455 이하 | 측정 불가(본문 블록 없음) |
| B2 | `case at 456:2` | case journal.StateConfirmed: | 아래 「State mutations」 | 소스 줄 456 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `case at 458:2` | case journal.StateFailedConfirmed, journal.StateNotDispatched: | 아래 「State mutations」 | 소스 줄 458 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `if at 460:3` | if detail == "" && err != nil { | 아래 「State mutations」 | 소스 줄 460 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `case at 464:2` | case journal.StateInDoubt, journal.StateUnresolvedInDoubt: | 아래 「State mutations」 | 소스 줄 464 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 467:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 467 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

없음(`err.Error`).

AST 호출 이름 전체: `err.Error`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
