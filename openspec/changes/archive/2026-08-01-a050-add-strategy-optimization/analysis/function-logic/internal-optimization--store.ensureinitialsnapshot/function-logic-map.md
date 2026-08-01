# Function Logic Map: `Store.ensureInitialSnapshot`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| control singleton | authenticated empty pointer or positive pointer bound to verified snapshot | v3 control/snapshot tables | corruption fails before initialization/readiness |
| initial values | only registry default/effective entries explicitly in value state | owner descriptors | no invented option for unapproved/read-only state |
| initial CAS | empty authenticated pointer to version 1 + exact snapshot digest | transaction | race loses cleanly and commits nothing partial |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | transaction/control read failure | none | error | DB fault coverage |
| B3-B5 | existing version path: snapshot or pointer digest invalid | read only | error; valid commits | pointer tamper/reopen tests |
| B6 | empty pointer digest invalid | none | corrupt pointer error | empty-pointer tamper test |
| B7-B9 | iterate registry; copy only explicit desired/effective value defaults | local maps | continue | initial snapshot tests |
| B10 | snapshot insert fails | transaction-local | error | fault coverage |
| B11 | authenticated control CAS update fails | transaction-local | error | fault coverage |
| B12 | CAS changed zero rows | rollback transaction | nil rollback result | concurrent initialization test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| snapshot read/digests | authenticate existing state | strict/no repair | pointer tests |
| registry iteration | derive owner-authored defaults | in-memory | initial state tests |
| `insertSnapshot` and control CAS | atomically establish version 1 | transaction rollback on failure/race | reopen/concurrency tests |

## State mutations and fallbacks

- Writes only an initial immutable snapshot and authenticated pointer within one transaction. Existing state is never reinitialized or selected by max version.

## Safety conclusion

- Safe edit boundary: authenticated idempotent store bootstrap.
- High-risk impact: yes; initial/current settings authority begins here.
