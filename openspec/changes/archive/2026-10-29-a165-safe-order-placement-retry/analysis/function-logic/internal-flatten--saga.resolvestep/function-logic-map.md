# Function Logic Map: `Saga.resolveStep`

- Source: `internal/flatten/flatten.go` (423-446)
- Qualified function: `Saga.resolveStep`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 1f66eed56b478d18…)
- AST evidence: `ast.json` — AST branches 6, 반환 5, 호출 7
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

flatten 의 IN_DOUBT 취소를 **동기로** 해소. B3 switch: CONFIRMED(434) / FAILED_CONFIRMED(437) / 그 밖(443). a165 R4 가 소진되면 결과가 오늘처럼 IN_DOUBT 여서 이 함수로 온다(1.3 재리뷰 F1 수리).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| step | FlattenStep | settleStep | B1 Resolver 없음 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 425:3, 430:3, 435:3, 441:3, 444:3

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 424:2` | if s.Resolver == nil \|\| step.AttemptID == "" { | 아래 「State mutations」 | 소스 줄 424 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 429:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 429 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `switch at 433:2` | switch res.State { | 아래 「State mutations」 | 소스 줄 433 이하 | 측정 불가(본문 블록 없음) |
| B4 | `case at 434:2` | case journal.StateConfirmed: | 아래 「State mutations」 | 소스 줄 434 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `case at 437:2` | case journal.StateFailedConfirmed: | 아래 「State mutations」 | 소스 줄 437 이하 | 미실행 — 패키지 스위트 count=0 |
| B6 | `case at 443:2` | default: | 아래 「State mutations」 | 소스 줄 443 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`s.Resolver.Resolve`.

AST 호출 이름 전체: `err.Error`, `s.Resolver.Resolve`, `string`.

## State mutations and fallbacks

- 없음(값 반환).

## Safety conclusion

- 편집 계획 없음.
