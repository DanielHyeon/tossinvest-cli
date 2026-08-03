# Function Logic Map: `evaluate`

- Source: `internal/continuationlane/evaluator.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed plan/evidence/cap/risk state and explicit caller-provided entry/target terms. No target inference is permitted.

## Branches and early returns

B1-B19 cover plan, invalidation, OFF, scope, signal, stop, leg, latch, cap, arithmetic and risk budget refusals. A new execution-term branch must require canonical `stop < entry < target` before decision.

## Calls and live bindings

Pure stop composition, allocation and risk arithmetic only.

## State mutations and fallbacks

None. Accepted outcome preserves canonical terms; refusal returns no terms.

## Safety conclusion

Safe edit boundary: explicit term validation immediately after effective stop composition. High-risk impact: malformed prices must never reach accepted output.
