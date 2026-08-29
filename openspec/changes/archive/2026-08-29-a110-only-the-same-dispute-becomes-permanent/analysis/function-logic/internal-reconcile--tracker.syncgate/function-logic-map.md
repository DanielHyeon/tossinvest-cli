# Function Logic Map: `Tracker.syncGate`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Authority | Failure behavior |
|---|---|---|---|
| active blocks | full current tracker projection | journal-confirmed plus pending memory | blank or unknown narrow scope projects conservatively account-wide |

## Branches and early returns

| Branches | Condition | Result | Test |
|---|---|---|---|
| B1 | no gate | no-op | nil gate tests |
| B2–B3 | account block or blank symbol | account reason projection | blank-symbol and permanent tests |
| B4–B5 | real symbol block | latch symbol | changing-symbol tests |
| B6–B8 | existing symbol is foreign reason/survives/disappears | preserve or exact clear | gate ownership tests |
| B9–B11 | each account reconcile reason present/absent | block or clear that reason only | permanent precedence tests |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `EntryGate.BlockSymbol/ClearSymbol` | narrow ordinary projection | other reason families untouched | B4–B8 |
| `EntryGate.Block/Clear` | account projection | permanent and ordinary reasons independent | B9–B11 |

## State mutations and fallbacks

- Mutates only the execution-gate projection. It does not change tracker/journal authority.
- Blank symbol is account-safe here and in `Block.Covers`, while remaining a distinct in-memory key.

## Safety conclusion

- Safe boundary: preserve reason ownership and permanent precedence; align blank-symbol projections.
- High-risk impact: yes; this is the actual entry interlock.
