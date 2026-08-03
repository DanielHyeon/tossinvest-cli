# Function Logic Map: `stopProvenance`

- Source: `internal/continuationlane/execution_terms.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan/envelope | sealed market plan and source evidence | evaluator request | invalid saved authority returns false |
| candidate/effective | effective price is candidate or monotonic saved stop | `effectiveStop` | unsupported scale or invalid seal returns false |
| saved | package-private, plan/evidence/price-bound sealed authority | `mintSavedStopProvenance` | zero/forged/mismatched seal returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | effective stop equals current candidate | none | candidate provenance | candidate execution-term tests |
| B2 | quote currency has no supported minor scale | none | empty provenance, false | unsupported-currency tests |
| fallthrough | saved stop selected | none | saved provenance only with valid private seal | `TestCallerForgedSavedStopProvenanceFailsClosed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currencyMinorScale` | bind provenance to canonical quote scale | false; no retry/fallback | AST B2 |
| `savedStopAuthority.valid` | verify private seal and plan/evidence/price binding | false; no retry/fallback | AST return |

## State mutations and fallbacks

- Pure composition. It does not mutate plan, evidence, candidate, or authority.
- There is no fallback from an invalid saved authority to caller-provided provenance.

## Safety conclusion

- Safe edit boundary: provenance selection after monotonic stop selection.
- High-risk impact: yes; all saved-stop provenance is fail-closed behind a package-private seal.
