# Function Logic Map: `Tracker.Resolve`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input | Valid range | Authority | Failure behavior |
|---|---|---|---|
| operator/note | both nonblank | human audited release | validation or atomic journal failure leaves all blocks/gates intact |

## Branches and early returns

| Branches | Condition | Result | Test |
|---|---|---|---|
| B1 | operator blank | reject | operator identity test |
| B2 | note blank | reject | operator note test |
| B3–B4 | journal present / each block | build atomic exact-cause release requests | operator release test |
| B5 | atomic release fails | return without memory clear | release failure test |
| Return | authority confirmed | clear permanent, streaks, pending identity, blocks, credits and gate | operator-only test |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `ReleaseReconciles` | atomic audited operator release | all-or-error | B5 |
| `clearStreaks`/`syncGate` | clear runtime only after authority | under lock | operator suite |

## State mutations and fallbacks

- This is the only automatic-code path that can clear durable permanent state, and it requires explicit human identity/note.

## Safety conclusion

- Safe boundary: no semantic release change in a110; map records preserved operator-only invariant.
- High-risk impact: yes; never invoke operationally without separate approval.
