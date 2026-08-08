# Function Logic Map: `TestEachInterlockClauseHasALine`

- Source: `internal/app/engine/runtime_wiring_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| clause cases | endpoint, limits, attestation and no-Guardian refusals | `UnmetInterlockClauses` | missing operator line fails |
| no-Guardian row | sealed test-only construction disable | interlock contract | production constructor would mask clause |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | iterate and choose gate writer | isolated setup | test failure |
| B4 | ordinary rows create matched Guardian | table data | continue |
| B5-B6 | no-Guardian row selects sealed helper, others normal assembly | choose path | continue |
| B7-B8 | require refusal and expected enumerated text | no mutation | test failure |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openGateWithoutProductionGuardian` | expose the specific clause | test-only | production unaffected |
| `engine.UnmetInterlockClauses` | render operator refusal list | exact text assertion | CodeGraph |

## State mutations and fallbacks

- No production safety behavior is disabled; the seam exists only in the test binary.

## Safety conclusion

- Safe edit boundary: select helper for the no-Guardian table row.
- High-risk impact: yes — operator refusal enumeration is a safety diagnostic.
