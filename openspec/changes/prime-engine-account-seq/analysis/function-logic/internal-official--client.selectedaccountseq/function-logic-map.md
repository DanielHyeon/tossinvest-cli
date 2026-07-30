# Function Logic Map: `Client.SelectedAccountSeq`

- Source: `internal/official/reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account sequence | zero unresolved, positive usable, negative invalid | `Client.accountSeq` | returns usable=false unless positive |
| atomic cache | lock-free load | `Client.accountSeq` | no wait behind account-list I/O |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | atomic leaf expression evaluates `accountSeq > 0` | none | `(sequence, usable)` | engine mismatch tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `accountSeq.Load` | race-free observation without discovery contention | no I/O or retry | post-edit AST |

## State mutations and fallbacks

- No account, network, journal, or cache mutation occurs.
- The method deliberately exposes negative values with `usable=false` so startup
  can report the mismatch without treating them as selected.

## Safety conclusion

- Safe edit boundary: atomic read only.
- High-risk impact: yes — engine startup uses this result to bind its journal
  identity to the request header scope.
