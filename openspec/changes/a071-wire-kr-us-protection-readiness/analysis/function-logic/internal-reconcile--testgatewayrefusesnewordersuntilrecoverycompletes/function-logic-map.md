# Function Logic Map: `TestGatewayRefusesNewOrdersUntilRecoveryCompletes`

- Source: `internal/reconcile/recovery_test.go` (420-538)
- Revision: base `775c37cb` (HEAD 에 함수 없음); source_sha256 `d5f2f4543c4663eb6f0830fa7b283f1292561f1dfe671b720b54b556ce53f8d9`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 13 · returns 1 · calls 43
- Exact AST return positions: 502:3
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | TestGatewayRefusesNewOrdersUntilRecoveryCompletes | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 439:2 | `if err != nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B2 | if at 444:2 | `if err != nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B3 | if at 467:3 | `if err != nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B4 | if at 477:3 | `if err != nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B5 | if at 480:3 | `if _, err := j.Reserve(context.Background(), journal.ReserveRequest{` | base 소스의 분기 — HEAD 에 함수 없음 |
| B6 | if at 508:2 | `if err == nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B7 | if at 511:2 | `if out.Reason != execgw.ReasonRecoveryIncomplete {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B8 | if at 514:2 | `if !strings.Contains(out.Detail, "recovery") {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B9 | if at 517:2 | `if broker.places != 0 {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B10 | if at 520:2 | `if out.State != journal.StateNotDispatched {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B11 | if at 525:2 | `if _, err := r.Run(context.Background()); err != nil {` | base 소스의 분기 — HEAD 에 함수 없음 |
| B12 | if at 530:2 | `if _, err := gw.Place(context.Background(), execgw.PlaceRequest{` | base 소스의 분기 — HEAD 에 함수 없음 |
| B13 | if at 535:2 | `if broker.places != 1 {` | base 소스의 분기 — HEAD 에 함수 없음 |

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

- Safe edit boundary: TestGatewayRefusesNewOrdersUntilRecoveryCompletes only.
- High-risk impact: reviewed and regression-tested.
