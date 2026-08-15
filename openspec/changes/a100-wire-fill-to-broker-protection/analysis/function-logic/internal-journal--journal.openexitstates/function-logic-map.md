# Function Logic Map: `Journal.OpenExitStates`

- Source: `internal/journal/apply_hook.go` (621-644)
- AST evidence: `ast.json` — AST branches 4.

## Inputs and invariants

This account-scoped query is both the observer working set and crash restoration definition. Its absence result is never a broker protection-cancel instruction.

## Branches and early returns

| Branch | Result |
|---|---|
| B1 | query failure returns |
| B2 | rows enumerate |
| B3 | scan failure discards partial set |
| B4 | iterator failure returns |

## Calls and live bindings

Calls query, `scanExitState`, and row error check; SELECT and Scan order must evolve together.

## State mutations and fallbacks

Read-only, no partial result fallback.

## Safety conclusion

Changing scan shape can halt every exit observer; changing membership can confuse a temporary absence with flatness.
