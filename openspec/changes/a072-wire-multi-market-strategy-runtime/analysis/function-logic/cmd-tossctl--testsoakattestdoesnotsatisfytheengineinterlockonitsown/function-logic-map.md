# Function Logic Map: `TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn`

- Source: `cmd/tossctl/soak_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake qualifying WTS soak record | complete shared read set, no supervised mutations | test fixture + engine required endpoints | generated attestation MUST still miss mutations but no shared GET |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | soak attest command errors | test files only | fatal | focused CLI test |
| B2 | attestation load errors | none | fatal | focused CLI test |
| B3 | no endpoint is missing | none | fatal because soak alone must not enable automation | focused CLI test |
| B4 | inspect each missing endpoint | none | continue through mutation gaps | focused CLI test |
| B5 | any shared GET is missing | none | test error | focused CLI test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `runCLI`, `attest.Load`, `MissingEndpoints` | build fake WTS evidence and compare engine contract | no real network; failures are test-fatal | CodeGraph + AST |

## State mutations and fallbacks

- Writes only isolated test evidence. It does not execute the official contract probe or mutation path.

## Safety conclusion

- Safe edit boundary: prove soak covers every shared read while supervised mutations remain outstanding.
- High-risk impact: **no** — test-only, but guards fail-closed startup evidence.
