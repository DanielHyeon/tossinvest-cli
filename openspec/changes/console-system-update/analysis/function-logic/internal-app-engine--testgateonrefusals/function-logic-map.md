# Function Logic Map: `TestGateOnRefusals`

- Source: `internal/app/engine/interlock_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| refusal table | each interlock clause and expected sentinel/text | engine safety spec | subtest fails on wrong clause |
| nil Guardian case | must disable automatic production construction only in test | test-only seam | would otherwise test successful wiring |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | iterate cases; optional trading/attestation setup | isolated config/broker | test setup failure |
| B4-B5 | no-Guardian case uses sealed disable seam; all others use normal helper | choose assembly | continue |
| B6-B10 | assert nil context, refusal sentinel/cause/text | no mutation | test failure |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openGateWithoutProductionGuardian` | retain direct missing-authority clause coverage | test-only | new Guardian production tests |
| `openGateEngine` | exercise ordinary interlock refusals | isolated broker | existing suite |

## State mutations and fallbacks

- Only the named no-Guardian row bypasses production construction.
- Other nil-Guardian production behavior is covered in the new assembly regressions.

## Safety conclusion

- Safe edit boundary: choose the sealed test helper for one table row.
- High-risk impact: yes — a false-positive interlock test could hide missing risk authority.
