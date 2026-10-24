# Function Logic Map: `Saga.settleStep`

- Source: `internal/flatten/flatten.go` (301-416)
- Qualified function: `Saga.settleStep`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 1f66eed56b478d18…)
- AST evidence: `ast.json` — AST branches 11, 반환 5, 호출 18
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

CLI flatten 의 취소 한 단계. B9(395) 분류가 InDoubt 면 B10(400) `resolveStep` 을 **동기로** 돈다(해소 최소 45s) — 409 를 확정 미수락으로 바꾸면 이 동기 해소를 타지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| step | FlattenStep | CancelAll | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 319:3, 331:3, 337:3, 379:3, 415:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `switch at 316:2` | switch step.State { | 아래 「State mutations」 | 소스 줄 316 이하 | 측정 불가(본문 블록 없음) |
| B2 | `case at 317:2` | case journal.FlattenStepDone, journal.FlattenStepFailed: | 아래 「State mutations」 | 소스 줄 317 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `case at 321:2` | case journal.FlattenStepInDoubt: | 아래 「State mutations」 | 소스 줄 321 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `if at 327:3` | if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State, | 아래 「State mutations」 | 소스 줄 327 이하 | 측정 불가(본문 블록 없음) |
| B5 | `if at 334:2` | if s.DryRun { | 아래 「State mutations」 | 소스 줄 334 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 348:2` | if target, ok := byOrder[step.TargetOrderID]; ok { | 아래 「State mutations」 | 소스 줄 348 이하 | 실행됨 — 패키지 스위트 count=1 |
| B7 | `if at 368:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 368 이하 | 미실행 — 패키지 스위트 count=0 |
| B8 | `if at 375:3` | if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State, | 아래 「State mutations」 | 소스 줄 375 이하 | 측정 불가(본문 블록 없음) |
| B9 | `if at 395:2` | if state == journal.FlattenStepInDoubt { | 아래 「State mutations」 | 소스 줄 395 이하 | 실행됨 — 패키지 스위트 count=1 |
| B10 | `if at 400:2` | if outcome.State == journal.FlattenStepInDoubt { | 아래 「State mutations」 | 소스 줄 400 이하 | 실행됨 — 패키지 스위트 count=1 |
| B11 | `if at 411:2` | if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State, | 아래 「State mutations」 | 소스 줄 411 이하 | 측정 불가(본문 블록 없음) |

## Calls and live bindings

`decisionFor`, `Gateway.Cancel`, `classifyCancel`, `resolveStep`, `UpdateFlattenStep`.

AST 호출 이름 전체: `classifyCancel`, `err.Error`, `parseDecimal`, `s.Gateway.Cancel`, `s.Journal.UpdateFlattenStep`, `s.decisionFor`, `s.event`, `s.reductionReason`, `s.resolveStep`, `string`.

## State mutations and fallbacks

- flatten step 상태.

## Safety conclusion

- 편집 계획 없음 — 배선(cmd/tossctl/flatten.go)만.
