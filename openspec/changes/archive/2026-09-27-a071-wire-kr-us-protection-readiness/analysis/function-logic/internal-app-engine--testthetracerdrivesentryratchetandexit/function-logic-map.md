# Function Logic Map: `TestTheTracerDrivesEntryRatchetAndExit`

- Source: `internal/app/engine/tracer_test.go` (116-144)
- Revision: base `775c37cb` (HEAD 에 함수 없음); source_sha256 `6d50eb9c3d64746ce4b3430c56a3b714fe00cf852325806b2bdc1cf73014e582`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 6 · returns 0 · calls 12
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | TestTheTracerDrivesEntryRatchetAndExit | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 126:2 | `if err != nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B2 | if at 129:2 | `if report.EntryOrderID == "" {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B3 | if at 132:2 | `if !report.Closed {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B4 | if at 135:2 | `if report.Proposals == 0 {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B5 | if at 138:2 | `if report.Outcome == nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B6 | if at 141:2 | `if report.Outcome.InitialQuantity != "1" {` | base 소스의 분기 — HEAD 에 함수 없음 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- Base-revision evidence records the removed or renamed scalar-test path.

## Safety conclusion

- Safe edit boundary: TestTheTracerDrivesEntryRatchetAndExit only.
- High-risk impact: reviewed and regression-tested.
