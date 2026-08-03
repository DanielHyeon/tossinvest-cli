# Function Logic Map: `fromWeekly`

- Source: `internal/strategyflow/adapters.go`
- AST evidence: `ast.json`

## Inputs and invariants

Adapts weekly outcome and plan account scope. Weekly target must be the evaluator's exact capped target.

## Branches and early returns

Branchless mapping; refusal never acquires execution terms.

## Calls and live bindings

Calls only pure weekly risk digest derivation.

## State mutations and fallbacks

None. It must not use staged or fair value inputs as a replacement for the evaluated capped target.

## Safety conclusion

Safe edit boundary: copy exact output terms. High-risk impact: target substitution must fail closed.
