# Function Logic Map: `fromReversal`

- Source: `internal/strategyflow/adapters.go`
- AST evidence: `ast.json`

## Inputs and invariants

Adapts a reversal decision plus exact evidence envelope scope.

## Branches and early returns

Branchless mapping; accepted terms must be copied byte-exactly.

## Calls and live bindings

No external calls or mutation capabilities.

## State mutations and fallbacks

None; missing terms are not synthesized.

## Safety conclusion

Safe edit boundary: add value fields only. High-risk impact: downstream execution term completeness.
