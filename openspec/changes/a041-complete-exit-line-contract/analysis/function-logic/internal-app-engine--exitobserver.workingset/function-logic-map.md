# Function Logic Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| positions and exit states | held, eligible, incomplete | journal projection | alert/skip unmanaged; return storage error |
| runtime policy identity | fixed legacy identity compatible with stored ID/adoption kind | exitpolicy compatibility registry | refuse ambiguous/unknown semantics |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | closed/zero position | none | skip | existing reconciliation tests |
| B2 | ineligible position | alert latch | skip | existing unmanaged test |
| B3 | missing exit state | open from entry/adoption seed | skip on completed/error | existing opening tests |
| B4 | state identity cannot be resolved | cycle refusal | skip position | identity conflict test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Positions`/`OpenExitStates` | reconcile held positions to policy rows | errors stop cycle | CodeGraph + AST |
| `openState` | create t0 protection | duplicate completed row is skipped | CodeGraph + AST |
| `LegacyPolicyIdentity` | bind ID-only pre-a042 row to one fixed meaning | unknown is fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- Adds an in-memory immutable identity to the managed read model; no a042 schema columns are claimed.

## Safety conclusion

- Safe edit boundary: resolve exact compatibility identity while constructing the cycle working set.
- High-risk impact: yes — identity refusal prevents an unsafe policy reinterpretation.
