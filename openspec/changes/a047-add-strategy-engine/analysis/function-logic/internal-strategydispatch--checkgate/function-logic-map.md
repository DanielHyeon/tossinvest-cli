# Function Logic Map: `checkGate`

- Source: `internal/strategydispatch/dispatch.go`
- Source SHA-256: `43b611261cd8e13ced3490f883c9d1201e1fcf4a128faf3bde39c5d7f27d58eb`
- CodeGraph callers/callees: initial and leased checks inside `dispatchValidated`
- AST evidence: `ast.json` (`B1` through `B11`)

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| decision lane/threshold/evidence provenance | exact equality to manifest binding | opaque DecisionRecord + activation snapshot | activation refusal |
| complete decision binding | exact 60-field snapshot | gate authority | activation refusal |
| order settings | manifest settings digest plus fixed LIMIT/KRW | server manifest | activation refusal |
| operational blockers | authority-owned booleans | lane/kill/protection/reconcile/scheduler/gate stores | first stable refusal in switch order |

## Branches and early returns

| Branch | Exact AST condition | Mutation/side effect | Return/error | Direct test |
|---|---|---|---|---|
| B1 | lane ID/version, source/constants, threshold or evidence differs from manifest binding | none | activation | activation mismatch row |
| B2 | full decision binding differs, settings digest differs, or order is not LIMIT/KRW | none | activation | 60/60 mutation plus three direct order rows |
| B3 | ordered operational switch is evaluated | none | B4 through B11 or success | full initial gate table |
| B4 | lane desired or effective is OFF | none | lane off | desired/effective rows |
| B5 | kill switch is ON | none | kill | kill row |
| B6 | protection is not wired | none | protection | protection row |
| B7 | reconciliation is unhealthy | none | reconcile | reconcile row |
| B8 | scheduler is invalid | none | scheduler | scheduler row |
| B9 | autostart is OFF | none | autostart | autostart row |
| B10 | gate is closed | none | gate | gate row |
| B11 | LIVE is not approved | none | live | LIVE row |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| `DecisionBinding` | exact full-record value conversion | activation refusal | 60/60 reflection test |

## State mutations and fallbacks

- Pure ordered predicate. It has no mutation, I/O, retry or fallback.
- Activation checks precede every operational blocker. Operational precedence is lane, kill, protection, reconcile, scheduler, autostart, gate, then LIVE.

## Safety conclusion

- No user input or live side effect. Callers must treat the first returned reason as authoritative and never skip later authority rechecks.
