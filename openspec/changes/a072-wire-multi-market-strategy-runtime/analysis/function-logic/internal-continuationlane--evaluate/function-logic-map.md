# Function Logic Map: `evaluate`

- Source: `internal/continuationlane/evaluator.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed plan/evidence/cap/risk state and explicit caller-provided entry/target terms. No target inference is permitted.

## Branches and early returns

B1-B6 cover plan, invalidation, OFF, scope and signal refusals. B7 rejects a public saved-stop scalar without a valid private authority. B8-B20 cover stop, leg, latch, cap, arithmetic and risk budget refusals. B21 requires sealed exact execution terms and canonical `stop < entry < target` before decision. The selected saved price comes from private authority, so public scalar mutation cannot retreat it.

## Calls and live bindings

Private saved-stop price selection, pure stop composition, allocation and risk arithmetic only.

## State mutations and fallbacks

None. Accepted outcome preserves canonical terms. A valid private saved authority overrides the public compatibility scalar; an absent authority permits only an empty public scalar.

## Safety conclusion

Safe edit boundary: explicit term validation immediately after effective stop composition. High-risk impact: malformed prices must never reach accepted output.
