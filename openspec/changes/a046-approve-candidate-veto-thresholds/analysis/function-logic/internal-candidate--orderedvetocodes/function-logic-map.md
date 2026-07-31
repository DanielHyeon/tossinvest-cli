# Function Logic Map: `OrderedVetoCodes`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| private `orderedVetoCodes` | exactly seen_late, extended, near_high in D3 order | candidate package invariant | array value copy prevents external mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unconditional leaf | none | independent `[3]VetoCode` value | `TestOrderedVetoCodesReturnsAnExternallyMutableCopy` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct value return | no error, I/O, timeout, or retry | AST `calls: null` |

## State mutations and fallbacks

- No mutation and no fallback; callers cannot obtain the private array itself.

## Safety conclusion

- Safe edit boundary: expose fixed ordering without exported mutable state.
- High-risk impact: no; it tightens a read-only candidate invariant.
