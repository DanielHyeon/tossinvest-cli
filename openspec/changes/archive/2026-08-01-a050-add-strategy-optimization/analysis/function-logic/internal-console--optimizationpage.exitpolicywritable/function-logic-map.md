# Function Logic Map: `optimizationPage.ExitPolicyWritable`

- Source: `internal/console/optimization_view.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.LifecycleReady` | true only after successful commander read | console handler | false returns read-only |
| `p.Fields` | category-scoped registry projections | `optimizationFieldViews` | absent target key returns read-only |
| target field | exact key `exit.common-policy`, no configuration error, non-read-only control | owner registry | any mismatch returns read-only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | lifecycle is not ready | none | false | failed/unwired page has no forms |
| B2 | iterate category fields | none | continue or target result | complete and invalid descriptor tests |
| B3 | exact target key found | none | true only for valid writable owner descriptor | configuration error suppresses controls |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| field comparison only | fail-closed eligibility for preset form | pure/no retry | AST |

## State mutations and fallbacks

- Reads local page projection only; no registry/snapshot/LIVE mutation.
- No permissive fallback: missing seam, missing field, error, or read-only control all return false.

## Safety conclusion

- Safe edit boundary: only suppress or expose the existing server-option preview form.
- High-risk impact: presentation gate on a high-risk setting; fail-closed branches are mandatory.
