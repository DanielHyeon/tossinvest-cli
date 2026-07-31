# Function Logic Map: `TestAnAbsentThresholdIsNotAPassedVeto`

- Source: `internal/candidate/veto_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| measured observations, absent thresholds | thresholds all empty | D10/a046 dormant state | test fails if any code measures or passes |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | verdict passes | fatal test failure | stop | this test |
| B2 | iterate copied D3 order | test inspection | continue | this test |
| B3 | code measured | testing error | continue | this test |
| B4 | reason not threshold absent | testing error | continue | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AssessChase`, `OrderedVetoCodes`, state accessors | verify dormant fail-closed behavior | test-only | AST |

## State mutations and fallbacks

- Fixture only; no numeric fallback.

## Safety conclusion

- Safe edit boundary: expected-order accessor.
- High-risk impact: no; regression test protects fail-closed state.
