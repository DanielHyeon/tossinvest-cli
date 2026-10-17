# Function Logic Map: `TestEntryGateRefusesEntriesUntilRecoveryCompletes`

- Source: `internal/reconcile/recovery_test.go` (418-439)
- Revision: current — HEAD `648df8ef`; source_sha256 `28ddf9c8e7dbad7bec48eee2f540f5ec0a53257c25ec3daea65b854669545c2c`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 4 · returns 0 · calls 14
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | TestEntryGateRefusesEntriesUntilRecoveryCompletes | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 424:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 427:2 | `if rejected := gate.CheckEntry(); rejected == nil \|\| rejected.Reason != execgw.ReasonRecoveryIncomplete {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | if at 433:2 | `if _, err := r.Run(context.Background()); err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B4 | if at 436:2 | `if rejected := gate.CheckEntry(); rejected != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- No authority broadening; current behavior is covered by focused tests.

## Safety conclusion

- Safe edit boundary: TestEntryGateRefusesEntriesUntilRecoveryCompletes only.
- High-risk impact: reviewed and regression-tested.
