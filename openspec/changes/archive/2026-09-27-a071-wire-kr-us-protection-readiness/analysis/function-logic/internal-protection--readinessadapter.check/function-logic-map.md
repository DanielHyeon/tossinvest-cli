# Function Logic Map: `ReadinessAdapter.Check`

- Source: `internal/protection/readiness_adapter.go` (118-155)
- Revision: current — HEAD `648df8ef`; source_sha256 `fed31b776d08c46f1cce728bb81740eee99734d3a489e6a4050df87c6e70af2a`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 9 · returns 7 · calls 7
- Exact AST return positions: 120:3, 129:3, 132:3, 136:3, 147:3, 152:3, 154:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request | KR/US, exact entry order type and canonical positive integral quantity | execgw mutation plan | typed refusal before transport |
| supervisor contract | sealed per-market session/trigger/replace/capability binding | engine production supervisor | corrupt/missing contract fails closed |
| prior checkpoint | empty on first check or exact same market generation/identity | first readiness decision | drift returns typed refusal |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 119:2 | `if adapter == nil \|\| adapter.provider == nil \|\| adapter.seal != adapterSeal(adapter) {` | NOT entered 0/48 |
| B2 | switch at 123:2 | `switch request.Market {` | evaluated 2/48 |
| B3 | case at 124:2 | `case "kr":` | entered 1/48 |
| B4 | case at 126:2 | `case "us":` | entered 1/48 |
| B5 | case at 128:2 | `default:` | entered 1/48 |
| B6 | if at 131:2 | `if request.OrderType == "" \|\| request.Quantity == 0 {` | NOT entered 0/48 |
| B7 | if at 135:2 | `if err != nil {` | entered 1/48 |
| B8 | if at 146:2 | `if !decision.Allowed {` | entered 1/48 |
| B9 | if at 151:2 | `if previous.Valid() && (previous.market != checkpoint.market \|\| previous.generation != checkpoint.generat...` | NOT entered 0/48 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | adapter/provider/seal invalid | none | state-corrupt refusal | adapter corruption test |
| N2 | unsupported market or invalid order/quantity | none | invalid refusal | substitution matrix |
| N3 | provider error | none | provider-unavailable refusal | existing provider failure test |
| N4 | dispatch refuses exact merged scope | none | propagated typed refusal | dispatch substitution matrix |
| N5 | prior checkpoint differs | none | snapshot-drift refusal | drift test |
| N6 | exact current snapshot | none | sealed checkpoint | valid test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SnapshotProvider.Current` | obtain current immutable KR/US view | single read; error is fail-closed | current HEAD |
| `ReadinessSnapshot.Dispatch` | compare plan plus sealed supervisor contract | pure/no retry | current HEAD |

## State mutations and fallbacks

- No broker mutation, retry or toggle write. The adapter only narrows authority.

## Safety conclusion

- Safe edit boundary: merge plan-supplied fields with already sealed supervisor-only fields; never invent defaults.
- High-risk impact: yes — called twice around durable dispatch.
