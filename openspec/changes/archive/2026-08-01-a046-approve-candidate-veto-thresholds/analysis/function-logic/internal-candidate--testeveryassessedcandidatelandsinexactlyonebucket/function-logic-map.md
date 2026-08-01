# Function Logic Map: `TestEveryAssessedCandidateLandsInExactlyOneBucket`

- Source: `internal/candidate/veto_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| clear, raised, and unmeasured chase fixtures | one of each bucket plus mixed cases | `TallyVetoes` partition contract | test fails on total/bucket/code/reason drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | total mismatch | testing failure | continue | this test |
| B2 | passed mismatch | testing failure | continue | this test |
| B3 | vetoed mismatch | testing failure | continue | this test |
| B4 | unmeasured mismatch | testing failure | continue | this test |
| B5 | partition sum mismatch | testing failure | continue | this test |
| B6 | near-high missing count mismatch | testing failure | continue | this test |
| B7 | seen-late raised count mismatch | testing failure | continue | this test |
| B8 | iterate copied D3 order | test inspection | continue | this test |
| B9 | Raised key missing | testing failure | continue | this test |
| B10 | NotMeasured key missing | testing failure | continue | this test |
| B11 | reason count mismatch | testing failure | return | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `TallyVetoes`, `OrderedVetoCodes` | assert exclusive partition and complete keys | test-only | AST |

## State mutations and fallbacks

- Fixture and testing state only.

## Safety conclusion

- Safe edit boundary: expected code list comes from immutable copy accessor.
- High-risk impact: no; regression test covers tally invariants.
