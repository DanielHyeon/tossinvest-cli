# Function Logic Map: `TestSoakAndLiveEndpointsCoverTheEngineInterlock`

- Source: `cmd/tossctl/soak_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| soak plus supervised endpoint definitions | all WTS reads and mutations | soak/verifylive catalogs | only exact inactive official OAuth exchange read may remain uncovered |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | collect existing soak/live coverage | test-local map mutation | complete WTS coverage map | focused CLI test |
| B2 | engine endpoint is covered | none | continue | focused CLI test |
| B3 | endpoint is exact exchange read or unexpected gap | none | accept exchange-only gap; error otherwise | focused CLI test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `soak.RequiredEndpoints`, `soak.LiveOnlyEndpoints`, `engine.RequiredEndpoints` | compare catalogs without network | pure definitions | CodeGraph + AST |

## State mutations and fallbacks

- Test-local map only. Official schema contract does not masquerade as executed attestation evidence.

## Safety conclusion

- Safe edit boundary: require the one explicitly documented OAuth gap and reject every other drift.
- High-risk impact: **no** — test-only operational coverage assertion.
