# Function Logic Map: `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`

- Source: `cmd/tossctl/soak_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake WTS soak plus supervised mutation evidence | complete shared engine-start evidence | isolated fixtures | no global endpoint gap allowed; official FX remains strategy-local |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | soak attest command errors | isolated files only | fatal | focused CLI test |
| B2 | attestation load errors | none | fatal | focused CLI test |
| B3 | any shared startup endpoint remains missing | none | fatal | focused CLI test |
| B4 | supervised evidence count differs | none | fatal | focused CLI test |
| B5 | supervised proof source is empty | none | test error | focused CLI test |
| B6 | generated evidence claims official exchange without OAuth runner | none | test error | focused CLI test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `runCLI`, `attest.Load`, `MissingEndpoints` | generate/read fake evidence and compare exact gap | no real network; fail test on drift | CodeGraph + AST |

## State mutations and fallbacks

- Writes isolated fake records only. It cannot produce OAuth exchange evidence or enable LIVE automation.

## Safety conclusion

- Safe edit boundary: preserve mutation proof checks while keeping strategy-only FX out of startup evidence.
- High-risk impact: **no** — test-only but startup-safety relevant.
