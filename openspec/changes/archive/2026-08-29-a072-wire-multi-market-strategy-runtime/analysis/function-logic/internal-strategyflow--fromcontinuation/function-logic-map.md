# Function Logic Map: `fromContinuation`

- Source: `internal/strategyflow/adapters.go`
- AST evidence: `ast.json`

## Inputs and invariants

Adapts a continuation `Outcome`; accepted means exact decision kind and empty refusal code.

## Branches and early returns

Branchless mapping. The happy path must copy exact validated entry, effective stop and target without defaults or conversion.

## Calls and live bindings

No calls or external capabilities; lane-to-flow value adapter only.

## State mutations and fallbacks

None. Missing terms remain missing for typed flow refusal.

## Safety conclusion

Safe edit boundary: copy terms only. High-risk impact: low locally, high downstream if a field is omitted.
