# Function Logic Map: `TestRemovingAVetoCodeCannotRemoveItsVeto`

- Source: `internal/candidate/veto_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| returned D3 order copy | three exact codes | private `orderedVetoCodes` | mutation of copy must not change predicates/tally |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | order length differs | fatal test failure | stop | this test |
| B2 | iterate expected codes | test inspection | continue | this test |
| B3 | code/order mismatch | testing failure | continue | this test |
| B4 | raised chase passes after copy mutation | testing failure | continue | this test |
| B5 | raised chase not vetoed | testing failure | continue | this test |
| B6 | tally loses raised veto | testing failure | continue | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, chase predicates, `TallyVetoes` | prove copy isolation across consumers | test-only | AST |

## State mutations and fallbacks

- Mutates only the returned array copy; package state must remain unchanged.

## Safety conclusion

- Safe edit boundary: convert old shared mutation regression into copy-isolation regression.
- High-risk impact: no; test protects candidate veto integrity.
