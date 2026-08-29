# Function Logic Map: `qFinalKRRequest`

- Source: `internal/execgw/riskguardian_qfinal_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test rig/suffix/candidate | valid fixed KR scope and positive test authority | test fixture only | helper fails test on provenance constructor error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | five required dimensions | append exact bucket/reference fixtures | return complete request | all q_final tests |
| B2-B3 | policy/snapshot provenance constructor failure | test abort | t.Fatal | fixture integrity |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| execgw test-only FX binder | inject private-field opaque-equivalent proof unavailable in production | test binary only | new bypass tests |

## State mutations and fallbacks

- Existing fixture DTO remains, but test authority is attached through export_test private state.

## Safety conclusion

- Safe edit boundary: test helper cannot create a production API for arbitrary FX.
- High-risk impact: no production impact; verifies high-risk path.
