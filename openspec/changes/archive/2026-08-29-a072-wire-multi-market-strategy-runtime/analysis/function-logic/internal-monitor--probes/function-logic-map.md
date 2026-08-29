# Function Logic Map: `Probes`

- Source: `internal/monitor/probes.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| internal WTS operation catalog | read-only WTS probes only | `internal/ops` registry plus explicit WTS CLI probes | malformed registry is visible to stable-name/schema tests |
| planned official FX contract probe | separate function and official origin | official Open API exchange-rate contract | MUST NOT enter this WTS session registry |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | operation catalog contains zero or more probes | append each probe without changing its check | stable ordered list | registry tests |
| B2 | quote stock info status is not 200 | none | sanitized status error | privacy/schema tests |
| B3 | quote symbol path is absent/wrong type | none | sanitized path error | privacy/schema tests |
| tail | all explicit read-only WTS probes appended | none | complete registry | stable-name test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ops.NewCatalog().Probes` | reuse operation-owned WTS probes | pure registry construction | CodeGraph + AST |
| `statusAndPath`, `expectStatus`, `expectPath` | bind minimal schema invariants | sanitized error; no response values | existing tests |

## State mutations and fallbacks

- Allocates and returns a new slice; no network, credential, journal, broker, or mutation call.
- The official OAuth contract probe remains in `OfficialReadContractProbes`, outside this function.

## Safety conclusion

- Safe edit boundary: preserve WTS-only contents while an explicit separation test prevents auth-domain mixing.
- High-risk impact: **no** — read-only registry construction.
